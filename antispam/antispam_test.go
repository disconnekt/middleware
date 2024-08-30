package antispam

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWrap_AllowsRequests(t *testing.T) {
	allowed := true

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	blockFunc := func(w http.ResponseWriter, r *http.Request) {
		allowed = false
		w.WriteHeader(http.StatusTooManyRequests)
	}

	handler := Wrap(nextHandler, blockFunc)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	req.Header.Set("X-Real-Ip", "127.0.0.1")
	rec := httptest.NewRecorder()

	for i := 0; i < 3; i++ {
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rec.Code)
		}
		if !allowed {
			t.Fatal("Request was blocked unexpectedly")
		}
	}
}

func TestWrap_BlocksAfterLimit(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	blockFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}

	handler := Wrap(nextHandler, blockFunc)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	req.Header.Set("X-Real-Ip", "127.0.0.1")
	rec := httptest.NewRecorder()

	for i := 0; i < 4; i++ {
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if i < 3 && rec.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", rec.Code)
		}
		if i == 3 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("Expected status 429 (Too Many Requests), got %d", rec.Code)
		}
	}
}

func TestWrap_ClearsExpiredEntries(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	blockFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}

	handler := Wrap(nextHandler, blockFunc)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	req.Header.Set("X-Real-Ip", "127.0.0.1")
	rec := httptest.NewRecorder()

	for i := 0; i < 3; i++ {
		handler.ServeHTTP(rec, req)
	}

	time.Sleep((BlockTime*5)*time.Second + time.Millisecond*10)

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 after expiration, got %d", rec.Code)
	}
}
