package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"umrah-backend/internal/entity"
	"umrah-backend/internal/handler"
	"umrah-backend/internal/middleware"
	"umrah-backend/internal/repository"
	"umrah-backend/internal/service"
	"umrah-backend/internal/worker"
	"umrah-backend/pkg/database"
	"umrah-backend/pkg/notification"
	"umrah-backend/pkg/queue"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 0. Safety Check: Create upload directory
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	// 1. Connect DB, Redis & RabbitMQ
	db := database.ConnectPostgres()
	redisClient := database.ConnectRedis()
	rabbit := queue.ConnectRabbitMQ("amqp://guest:guest@rabbitmq:5672/")
	defer rabbit.Close()

	// 2. Auto Migrate
	log.Println("Migrating database...")
	db.AutoMigrate(
		&entity.User{},
		&entity.TravelPackage{},
		&entity.Booking{},
		&entity.Group{},
		&entity.GroupMember{},
		&entity.Message{},
		&entity.Itinerary{},
		&entity.Attendance{},
		&entity.Product{},
		&entity.Order{},
		&entity.Manasik{},
	)

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	chatRepo := repository.NewChatRepository(db)
	itineraryRepo := repository.NewItineraryRepository(db)
	commerceRepo := repository.NewCommerceRepository(db)
	pkgRepo := repository.NewPackageRepository(db)
	manasikRepo := repository.NewManasikRepository(db)

	// 4. [FIXED] Initialize FCM Service FIRST (Needed for Worker)
	fcmSvc := notification.NewFCMService("firebase-credentials.json")

	// 5. [FIXED] Setup Worker (Now fcmSvc exists)
	worker := worker.NewChatWorker(rabbit, chatRepo, groupRepo, fcmSvc)
	worker.Start()

	// 6. Initialize Services
	authSvc := service.NewAuthService(userRepo, redisClient)
	groupSvc := service.NewGroupService(groupRepo)
	trackingSvc := service.NewTrackingService(redisClient, userRepo)
	chatSvc := service.NewChatService(chatRepo, redisClient, rabbit)
	itinerarySvc := service.NewItineraryService(itineraryRepo)
	commerceSvc := service.NewCommerceService(commerceRepo)
	pkgSvc := service.NewPackageService(pkgRepo)
	manasikSvc := service.NewManasikService(manasikRepo)

	// 7. Initialize Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	trackingHandler := handler.NewTrackingHandler(trackingSvc)
	chatHandler := handler.NewChatHandler(chatSvc, groupRepo)
	itineraryHandler := handler.NewItineraryHandler(itinerarySvc)
	commerceHandler := handler.NewCommerceHandler(commerceSvc)
	pkgHandler := handler.NewPackageHandler(pkgSvc)
	manasikHandler := handler.NewManasikHandler(manasikSvc)

	// 8. Setup Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit
	})

	// 9. Global Middlewares
	app.Use(recover.New()) // Anti Crash
	app.Use(logger.New())

	// [NEW] Input Sanitization (Before other logic)
	app.Use(middleware.XSSSanitizer)

	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	// Serve Static Files
	app.Static("/uploads", "./uploads")

	// --- ROUTING ---
	api := app.Group("/api")

	// A. PUBLIC ROUTES
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)
	api.Get("/packages", pkgHandler.GetList)
	api.Get("/manasik", manasikHandler.GetList)

	// B. PROTECTED ROUTES (User Logged In)
	api.Use(middleware.Protected())                     // Check JWT Signature
	api.Use(middleware.CheckSingleSession(redisClient)) // Check Redis Session

	// 1. Group & Member
	api.Post("/groups/join", groupHandler.Join)
	api.Get("/groups/:id/members", groupHandler.GetMembers)

	// 2. Chat
	api.Get("/groups/:group_id/chat", chatHandler.GetHistory)
	api.Delete("/groups/:group_id/chat/:message_id", chatHandler.DeleteMessage)

	// 3. Tracking
	api.Get("/groups/:group_id/locations", trackingHandler.GetLocations)

	// 4. Itinerary & Attendance
	api.Get("/groups/:group_id/rundown", itineraryHandler.GetRundown)
	api.Post("/attendance/scan", itineraryHandler.Scan)

	// 5. Commerce
	api.Get("/products", commerceHandler.GetCatalog)
	api.Post("/orders", commerceHandler.CreateOrder)
	api.Get("/orders/my", commerceHandler.GetMyOrders)
	api.Post("/orders/:id/proof", commerceHandler.UploadProof)
	api.Post("/bookings", pkgHandler.Book)

	// C. ADMIN / MUTAWWIF ROUTES (RBAC)
	admin := api.Group("/admin", middleware.AuthorizeRole("ADMIN", "MUTAWWIF"))

	admin.Post("/groups", groupHandler.Create)
	admin.Post("/packages", pkgHandler.Create)
	admin.Post("/products", commerceHandler.CreateProduct)
	admin.Post("/manasik", manasikHandler.Create)
	admin.Post("/itineraries", itineraryHandler.Create)
	admin.Get("/itineraries/:id/attendance", itineraryHandler.GetReport)
	admin.Patch("/orders/:id/verify", commerceHandler.VerifyOrder)

	// --- WEBSOCKET ROUTE ---
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket Auth (Query Param)
	app.Use("/ws", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		TokenLookup: "query:token",
	}))

	app.Get("/ws/tracking/:group_id", websocket.New(trackingHandler.StreamLocation))
	app.Get("/ws/chat/:group_id", websocket.New(chatHandler.StreamChat))

	// 10. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// 11. Start Server
	log.Println("Server running on port :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Panic(err)
	}
}
