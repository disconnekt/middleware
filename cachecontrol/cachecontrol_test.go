package cachecontrol

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	app.Use(Wrap(func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}))

	req := httptest.NewRequest("GET", "http://localhost/", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)

	assert.Equal(t, "Accept-Encoding", resp.Header.Get("Vary"), "Header 'Vary' is incorrect.")
	assert.Equal(t, "public, max-age=7776000", resp.Header.Get("Cache-Control"), "Header 'Cache-Control' is incorrect.")
}
