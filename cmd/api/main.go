package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"umrah-backend/internal/entity"
	"umrah-backend/internal/handler"
	"umrah-backend/internal/middleware"
	"umrah-backend/internal/repository"
	"umrah-backend/internal/service"
	"umrah-backend/pkg/database"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. Connect DB & Redis
	db := database.ConnectPostgres()
	redisClient := database.ConnectRedis()

	// 2. Auto Migrate (All Tables)
	db.AutoMigrate(
		&entity.User{},
		&entity.Group{},
		&entity.GroupMember{},
		&entity.Message{},
		&entity.Itinerary{},
		&entity.Attendance{},
		&entity.Product{},
		&entity.Order{},
		// [NEW]
		&entity.TravelPackage{},
		&entity.Booking{},
		&entity.Manasik{},
	)

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	chatRepo := repository.NewChatRepository(db)
	itineraryRepo := repository.NewItineraryRepository(db)
	commerceRepo := repository.NewCommerceRepository(db)
	// [NEW]
	pkgRepo := repository.NewPackageRepository(db)
	manasikRepo := repository.NewManasikRepository(db)

	// 4. Initialize Services
	authSvc := service.NewAuthService(userRepo, redisClient)
	groupSvc := service.NewGroupService(groupRepo)
	trackingSvc := service.NewTrackingService(redisClient, userRepo)
	chatSvc := service.NewChatService(chatRepo, redisClient)
	itinerarySvc := service.NewItineraryService(itineraryRepo)
	commerceSvc := service.NewCommerceService(commerceRepo)
	// [NEW]
	pkgSvc := service.NewPackageService(pkgRepo)
	manasikSvc := service.NewManasikService(manasikRepo)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	trackingHandler := handler.NewTrackingHandler(trackingSvc)
	chatHandler := handler.NewChatHandler(chatSvc, groupRepo)
	itineraryHandler := handler.NewItineraryHandler(itinerarySvc)
	commerceHandler := handler.NewCommerceHandler(commerceSvc)
	// [NEW]
	pkgHandler := handler.NewPackageHandler(pkgSvc)
	manasikHandler := handler.NewManasikHandler(manasikSvc)

	// 6. Setup Fiber
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format:     `{"time":"${time}","status":${status},"method":"${method}","path":"${path}"}` + "\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(helmet.New())
	app.Use(cors.New())
	app.Use(limiter.New(limiter.Config{Max: 100, Expiration: 1 * time.Minute}))

	app.Static("/uploads", "./uploads")

	api := app.Group("/api")

	// Public Routes
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)

	// Public Catalog (Can be accessed without login if desired, or move to protected)
	api.Get("/packages", pkgHandler.GetList)

	// --- PROTECTED ROUTES ---
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
	}))
	api.Use(middleware.CheckSingleSession(redisClient))

	// Group
	api.Post("/groups", groupHandler.Create)
	api.Post("/groups/join", groupHandler.Join)
	api.Get("/groups/:id/members", groupHandler.GetMembers)

	// Chat
	api.Get("/groups/:group_id/chat", chatHandler.GetHistory)
	api.Delete("/groups/:group_id/chat/:message_id", chatHandler.DeleteMessage)

	// Tracking
	api.Get("/groups/:group_id/locations", trackingHandler.GetLocations)

	// Itinerary & Attendance
	api.Post("/itineraries", itineraryHandler.Create)
	api.Get("/groups/:group_id/rundown", itineraryHandler.GetRundown)
	api.Post("/attendance/scan", itineraryHandler.Scan)
	api.Get("/itineraries/:id/attendance", itineraryHandler.GetReport)

	// Commerce (Roaming)
	api.Post("/products", commerceHandler.CreateProduct)
	api.Get("/products", commerceHandler.GetCatalog)
	api.Post("/orders", commerceHandler.CreateOrder)
	api.Get("/orders/my", commerceHandler.GetMyOrders)
	api.Post("/orders/:id/proof", commerceHandler.UploadProof)
	api.Patch("/orders/:id/verify", commerceHandler.VerifyOrder)

	// [NEW] Travel Packages & Booking
	api.Post("/packages", pkgHandler.Create) // Admin only check inside handler ideally
	api.Post("/bookings", pkgHandler.Book)

	// [NEW] Manasik
	api.Post("/manasik", manasikHandler.Create) // Admin only
	api.Get("/manasik", manasikHandler.GetList)

	// WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Use("/ws", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		TokenLookup: "query:token",
	}))

	app.Get("/ws/tracking/:group_id", websocket.New(trackingHandler.StreamLocation))
	app.Get("/ws/chat/:group_id", websocket.New(chatHandler.StreamChat))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_ = <-c
		log.Println("Shutting down...")
		_ = app.Shutdown()
	}()

	log.Println("Server running on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Panic(err)
	}
}
