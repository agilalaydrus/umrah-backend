package main

import (
	"log"
	"os"

	"umrah-backend/internal/entity"
	"umrah-backend/internal/handler"
	"umrah-backend/internal/repository"
	"umrah-backend/internal/service"
	"umrah-backend/pkg/database"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. Connect Database & Redis
	db := database.ConnectPostgres()
	redisClient := database.ConnectRedis()

	// --- AUTO MIGRATE (Updated for Sprint 4) ---
	log.Println("Running Database Migration...")
	err := db.AutoMigrate(
		&entity.User{},
		&entity.Group{},
		&entity.GroupMember{},
		&entity.Message{}, // <--- BARU: Tabel Chat
	)
	if err != nil {
		log.Fatal("Migration failed: ", err)
	}
	log.Println("Database Migration Success!")
	// -------------------------------------------

	// 2. Setup Dependencies
	// Auth & Group
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authSvc)

	groupRepo := repository.NewGroupRepository(db)
	groupSvc := service.NewGroupService(groupRepo)
	groupHandler := handler.NewGroupHandler(groupSvc)

	// Tracking Module
	trackingSvc := service.NewTrackingService(redisClient, userRepo)
	trackingHandler := handler.NewTrackingHandler(trackingSvc)

	// --- CHAT MODULE (BARU) ---
	chatRepo := repository.NewChatRepository(db)
	chatSvc := service.NewChatService(chatRepo, redisClient)
	chatHandler := handler.NewChatHandler(chatSvc)
	// --------------------------

	// 3. Setup Fiber
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	// 4. Routes
	api := app.Group("/api")

	// Auth Public Routes
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)

	// --- PROTECTED HTTP ROUTES ---
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
	}))

	// Group Routes
	api.Post("/groups", groupHandler.Create)
	api.Post("/groups/join", groupHandler.Join)
	api.Get("/groups/:id/members", groupHandler.GetMembers)

	// Tracking Routes
	api.Get("/groups/:group_id/locations", trackingHandler.GetLocations)

	// Chat Routes (History)
	chatGroup := api.Group("/groups/:group_id/chat")
	// Get History
	chatGroup.Get("/history", chatHandler.GetHistory)
	// URL: DELETE /api/groups/{group_id}/chat/{message_id}
	chatGroup.Delete("/:message_id", chatHandler.DeleteMessage)

	// --- WEBSOCKET SETUP ---
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Middleware: Validasi Token dari URL Query (?token=...)
	app.Use("/ws", jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		TokenLookup: "query:token",
	}))

	// Tracking WebSocket
	app.Get("/ws/tracking/:group_id", websocket.New(trackingHandler.StreamLocation))

	// Chat WebSocket - BARU
	app.Get("/ws/chat/:group_id", websocket.New(chatHandler.StreamChat))

	// Start Server
	log.Fatal(app.Listen(":3000"))
}
