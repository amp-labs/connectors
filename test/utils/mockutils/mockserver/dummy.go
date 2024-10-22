package mockserver

import (
	"net/http"
	"net/http/httptest"
)

// Dummy server only talks about having a cup of tea.
// Acknowledges requests and does nothing else.
func Dummy() *httptest.Server {
	// This is a factory method. Every server instance will be deleted after the test suite finishes.
	return NewServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
}

// NewServer is syntactic sugar for creating test server.
func NewServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}
