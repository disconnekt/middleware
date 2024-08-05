package cachecontrol

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	t.Parallel()

	nh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testData := []struct {
			name string
			val  string
		}{
			{
				name: "Vary",
				val:  "Accept-Encoding",
			},
			{
				name: "Cache-Control",
				val:  "public, max-age=7776000",
			},
		}

		for _, td := range testData {
			assert.Equal(t, td.val, w.Header().Get(td.name), "Header is incorrect.")
		}
	})

	Wrap(nh).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://localhost/", nil))
}
