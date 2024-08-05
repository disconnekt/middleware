package antispam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	t.Parallel()

	const allowRequests = 3
	var i, j int

	nh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i++
		assert.LessOrEqual(t, i, allowRequests, "Antispam allow more than possible requests.")
	})

	h := Wrap(nh, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/thankyou.html")
		w.WriteHeader(http.StatusFound)
	})

	for j = 0; j < allowRequests+1; j++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "http://localhost/", nil))
	}

	assert.GreaterOrEqual(t, j, allowRequests+1, "Antispam not checked.")
}
