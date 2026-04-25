package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

const requestIDHeader = "X-Request-ID"

func getRequestID(c fiber.Ctx) string {
	if v := requestid.FromContext(c); v != "" {
		return v
	}

	if v := c.Get(requestIDHeader); v != "" {
		return v
	}

	return ""
}
