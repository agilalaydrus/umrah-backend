package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func CheckSingleSession(redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Get Token from Locals (Set by JWT Middleware)
		userToken := c.Locals("user").(*jwt.Token)
		claims := userToken.Claims.(jwt.MapClaims)

		userID := claims["user_id"].(string)
		tokenSessionID := claims["sid"].(string) // The ID inside this token

		// 2. Get Active Session from Redis
		redisKey := fmt.Sprintf("session:user:%s", userID)
		activeSessionID, err := redisClient.Get(c.Context(), redisKey).Result()

		if err == redis.Nil {
			// Key not found = Session expired or user forced logout
			return c.Status(401).JSON(fiber.Map{"error": "Session expired, please login again"})
		} else if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Session check failed"})
		}

		// 3. Compare
		// If the token's SID is different from Redis SID, it means
		// someone else logged in on another device.
		if tokenSessionID != activeSessionID {
			return c.Status(401).JSON(fiber.Map{
				"error": "You have been logged out because this account logged in on another device.",
			})
		}

		return c.Next()
	}
}
