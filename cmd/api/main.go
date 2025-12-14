package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"umrah-backend/internal/entity"
	"umrah-backend/internal/handler"
	"umrah-backend/internal/middleware" // [NEW] Import Middleware Package
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
	// 1. Connect DB (Pooled) & Redis
	db := database.ConnectPostgres()
	redisClient := database.ConnectRedis()

	// Migrations
	db.AutoMigrate(&entity.User{}, &entity.Group{}, &entity.GroupMember{}, &entity.Message{})

	// 2. Dependencies
	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	chatRepo := repository.NewChatRepository(db)

	// [UPDATE] Auth Service now gets Redis Client
	authSvc := service.NewAuthService(userRepo, redisClient)

	groupSvc := service.NewGroupService(groupRepo)
	trackingSvc := service.NewTrackingService(redisClient, userRepo)
	chatSvc := service.NewChatService(chatRepo, redisClient)

	authHandler := handler.NewAuthHandler(authSvc)
	groupHandler := handler.NewGroupHandler(groupSvc)
	trackingHandler := handler.NewTrackingHandler(trackingSvc)

	// [CRITICAL] Inject groupRepo into ChatHandler for Security Check
	chatHandler := handler.NewChatHandler(chatSvc, groupRepo)

	// 3. Fiber Setup (Production Ready)
	app := fiber.New()

	// A. JSON Logger
	app.Use(logger.New(logger.Config{
		Format:     `{"time":"${time}","status":${status},"latency":"${latency}","method":"${method}","path":"${path}","error":"${error}"}` + "\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta",
	}))

	// B. Security Headers
	app.Use(helmet.New())

	// C. Rate Limiting (Prevent DDoS / Abuse)
	app.Use(limiter.New(limiter.Config{
		Max:        100,             // 100 requests
		Expiration: 1 * time.Minute, // per 1 minute
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Too many requests"})
		},
	}))

	// D. CORS
	app.Use(cors.New())

	// 4. Routes
	api := app.Group("/api")

	// Public Routes
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)

	// --- PROTECTED ROUTES START ---
	// 1. JWT Verification (Standard)
	api.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
	}))

	// 2. [NEW] Single Session Enforcement
	// Checks if the session ID in the token matches the one in Redis
	api.Use(middleware.CheckSingleSession(redisClient))

	// Group Routes
	api.Post("/groups", groupHandler.Create)
	api.Post("/groups/join", groupHandler.Join)
	api.Get("/groups/:id/members", groupHandler.GetMembers)

	// Tracking Routes
	api.Get("/groups/:group_id/locations", trackingHandler.GetLocations)

	// Chat Routes (History & Delete)
	api.Get("/groups/:group_id/chat", chatHandler.GetHistory)
	api.Delete("/groups/:group_id/chat/:message_id", chatHandler.DeleteMessage)
	// --- PROTECTED ROUTES END ---

	// WebSocket Middleware
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

	// Note: You can also wrap WebSocket routes with Session Check if needed,
	// but usually handling it at handshake (HTTP Upgrade) is sufficient if wired correctly.
	// For strictness, you could check session in the Handler's initial connection logic.

	app.Get("/ws/tracking/:group_id", websocket.New(trackingHandler.StreamLocation))
	app.Get("/ws/chat/:group_id", websocket.New(chatHandler.StreamChat))

	// 5. Graceful Shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		_ = <-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	log.Println("Server running on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Panic(err)
	}
}
