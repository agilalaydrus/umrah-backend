package middleware

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/microcosm-cc/bluemonday"
)

// XSSSanitizer cleans all incoming JSON bodies
func XSSSanitizer(c *fiber.Ctx) error {
	// Only run on methods that have a body
	if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
		var body map[string]interface{}

		// 1. Parse the body
		if err := c.BodyParser(&body); err != nil {
			// If parsing fails (e.g. empty body), just continue
			return c.Next()
		}

		// 2. Sanitize recursively
		policy := bluemonday.UGCPolicy() // Strict policy (User Generated Content)
		sanitizeMap(body, policy)

		// 3. Re-encode and set back to request
		sanitizedJSON, _ := json.Marshal(body)
		c.Request().SetBody(sanitizedJSON)
	}
	return c.Next()
}

// Helper to handle nested maps and lists
func sanitizeMap(m map[string]interface{}, policy *bluemonday.Policy) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = policy.Sanitize(val)
		case map[string]interface{}:
			sanitizeMap(val, policy)
		case []interface{}:
			sanitizeSlice(val, policy)
		}
	}
}

func sanitizeSlice(s []interface{}, policy *bluemonday.Policy) {
	for i, v := range s {
		switch val := v.(type) {
		case string:
			s[i] = policy.Sanitize(val)
		case map[string]interface{}:
			sanitizeMap(val, policy)
		}
	}
}
