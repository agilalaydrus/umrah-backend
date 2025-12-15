package middleware

import (
	"fmt"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// ---------------------------------------------------------
// 1. JWT Protected (Base Middleware)
// ---------------------------------------------------------
// Middleware ini memvalidasi Signature Token & Expiry
func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or expired token",
			})
		},
	})
}

// ---------------------------------------------------------
// 2. Single Session Check (Redis)
// ---------------------------------------------------------
// Middleware ini memvalidasi Session ID di Redis
func CheckSingleSession(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// [SAFETY FIX 1] Pastikan token ada di Locals
		userToken, ok := c.Locals("user").(*jwt.Token)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: Token not found"})
		}

		// [SAFETY FIX 2] Pastikan claims valid
		claims, ok := userToken.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: Invalid claims"})
		}

		// [SAFETY FIX 3] Ambil data dengan aman (hindari panic jika key tidak ada)
		userID, okID := claims["user_id"].(string)
		tokenSID, okSID := claims["sid"].(string)

		if !okID || !okSID {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized: Token structure invalid"})
		}

		// 2. Cek Redis
		redisKey := fmt.Sprintf("session:user:%s", userID)
		activeSessionID, err := redisClient.Get(c.Context(), redisKey).Result()

		if err == redis.Nil {
			// Session di Redis hilang (Expired atau Logout)
			return c.Status(401).JSON(fiber.Map{"error": "Session expired, please login again"})
		} else if err != nil {
			// Error Redis (Down/Timeout)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error (session check)"})
		}

		// 3. Bandingkan Session ID Token vs Redis
		if tokenSID != activeSessionID {
			return c.Status(401).JSON(fiber.Map{
				"error": "Anda telah login di perangkat lain. Silakan login kembali.",
				"code":  "FORCE_LOGOUT", // Code khusus agar frontend bisa redirect ke login
			})
		}

		return c.Next()
	}
}

// ---------------------------------------------------------
// 3. Role Based Access Control (RBAC)
// ---------------------------------------------------------
// Middleware untuk membatasi akses berdasarkan Role
func AuthorizeRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)
		userRole := claims["role"].(string)

		for _, role := range allowedRoles {
			if role == userRole {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error": "Forbidden: You do not have access to this resource",
		})
	}
}
