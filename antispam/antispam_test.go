package antispam

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestWrap_AllowsRequests(t *testing.T) {
	allowed := true

	nextHandler := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}

	blockFunc := func(c *fiber.Ctx) error {
		allowed = false
		return c.SendStatus(fiber.StatusTooManyRequests)
	}

	handler := WrapFiber(nextHandler, blockFunc)

	app := fiber.New()
	app.Get("/foo", handler)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("X-Real-Ip", "127.0.0.1")
		resp, _ := app.Test(req)

		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
		if !allowed {
			t.Fatal("Request was blocked unexpectedly")
		}
	}
}

func TestWrap_BlocksAfterLimit(t *testing.T) {
	nextHandler := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}

	blockFunc := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTooManyRequests)
	}

	handler := WrapFiber(nextHandler, blockFunc)

	app := fiber.New()
	app.Get("/foo", handler)

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("X-Real-Ip", "127.0.0.1")
		resp, _ := app.Test(req)

		if i < 3 && resp.StatusCode != fiber.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}
		if i == 3 && resp.StatusCode != fiber.StatusTooManyRequests {
			t.Fatalf("Expected status 429 (Too Many Requests), got %d", resp.StatusCode)
		}
	}
}

func TestWrap_ClearsExpiredEntries(t *testing.T) {
	nextHandler := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}

	blockFunc := func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusTooManyRequests)
	}

	handler := WrapFiber(nextHandler, blockFunc)

	app := fiber.New()
	app.Get("/foo", handler)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		req.Header.Set("X-Real-Ip", "127.0.0.1")
		app.Test(req)
	}

	time.Sleep((BlockTime*5)*time.Second + time.Millisecond*10)

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("X-Real-Ip", "127.0.0.1")
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("Expected status 200 after expiration, got %d", resp.StatusCode)
	}
}
