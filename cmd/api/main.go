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
	"umrah-backend/pkg/database" // Pastikan package ini ada

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
	// 0. Safety Check: Buat folder uploads jika belum ada
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	// 1. Connect DB & Redis
	db := database.ConnectPostgres()
	redisClient := database.ConnectRedis()

	// 2. Auto Migrate (Urutan diperbaiki agar relasi aman)
	log.Println("Migrating database...")
	db.AutoMigrate(
		&entity.User{},
		&entity.TravelPackage{}, // Package dulu baru Booking
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

	// 4. Initialize Services
	authSvc := service.NewAuthService(userRepo, redisClient)
	groupSvc := service.NewGroupService(groupRepo)
	trackingSvc := service.NewTrackingService(redisClient, userRepo)
	chatSvc := service.NewChatService(chatRepo, redisClient)
	itinerarySvc := service.NewItineraryService(itineraryRepo)
	commerceSvc := service.NewCommerceService(commerceRepo)
	pkgSvc := service.NewPackageService(pkgRepo)
	manasikSvc := service.NewManasikService(manasikRepo)

	// 5. Initialize Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	trackingHandler := handler.NewTrackingHandler(trackingSvc)
	chatHandler := handler.NewChatHandler(chatSvc, groupRepo)
	itineraryHandler := handler.NewItineraryHandler(itinerarySvc)
	commerceHandler := handler.NewCommerceHandler(commerceSvc)
	pkgHandler := handler.NewPackageHandler(pkgSvc)
	manasikHandler := handler.NewManasikHandler(manasikSvc)

	// 6. Setup Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // Limit upload 10MB
	})

	// Global Middlewares
	app.Use(recover.New()) // Anti Crash
	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Ganti dengan domain frontend nanti
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))

	// Serve Static Files (Images)
	app.Static("/uploads", "./uploads")

	// --- ROUTING ---
	api := app.Group("/api")

	// A. PUBLIC ROUTES
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)
	api.Get("/packages", pkgHandler.GetList) // Katalog Publik
	api.Get("/manasik", manasikHandler.GetList)

	// B. PROTECTED ROUTES (User Logged In)
	// Menggunakan Middleware custom yang kita buat sebelumnya
	api.Use(middleware.Protected())                     // Cek Token Signature
	api.Use(middleware.CheckSingleSession(redisClient)) // Cek Redis Session

	// 1. Group & Member
	api.Post("/groups/join", groupHandler.Join)
	api.Get("/groups/:id/members", groupHandler.GetMembers)

	// 2. Chat (History & Delete)
	api.Get("/groups/:group_id/chat", chatHandler.GetHistory)
	api.Delete("/groups/:group_id/chat/:message_id", chatHandler.DeleteMessage)

	// 3. Tracking
	api.Get("/groups/:group_id/locations", trackingHandler.GetLocations) // Snapshot API

	// 4. Itinerary & Attendance
	api.Get("/groups/:group_id/rundown", itineraryHandler.GetRundown)
	api.Post("/attendance/scan", itineraryHandler.Scan) // Jamaah Scan QR

	// 5. Commerce (User Side)
	api.Get("/products", commerceHandler.GetCatalog)
	api.Post("/orders", commerceHandler.CreateOrder)
	api.Get("/orders/my", commerceHandler.GetMyOrders)
	api.Post("/orders/:id/proof", commerceHandler.UploadProof)
	api.Post("/bookings", pkgHandler.Book)

	// C. ADMIN / MUTAWWIF ROUTES (RBAC)
	// Kita buat group khusus admin agar lebih aman
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

	// WebSocket Auth (Token lewat Query Param: ws://...?token=xyz)
	app.Use("/ws", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		TokenLookup: "query:token",
	}))

	app.Get("/ws/tracking/:group_id", websocket.New(trackingHandler.StreamLocation))
	app.Get("/ws/chat/:group_id", websocket.New(chatHandler.StreamChat))

	// 7. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	log.Println("Server running on port :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Panic(err)
	}
}
