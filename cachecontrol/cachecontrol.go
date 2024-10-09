package cachecontrol

import "github.com/gofiber/fiber/v2"

func Wrap(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Vary", "Accept-Encoding")
		c.Set("Cache-Control", "public, max-age=7776000")
		return next(c)
	}
}
