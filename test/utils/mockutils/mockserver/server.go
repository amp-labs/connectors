package mockserver

import (
	"net/http"
	"net/http/httptest"
)

type ServerBuilder struct {
	withJSON       bool
	responseStatus int
	responseBody   []byte
}

// New creates new builder for Mock Server.
func New() *ServerBuilder {
	return &ServerBuilder{}
}

// Build will finalize mock server.
// It returns the configured httptest.Server.
func (s *ServerBuilder) Build() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.withJSON {
			w.Header().Set("Content-Type", "application/json")
		}

		if s.responseStatus != 0 {
			w.WriteHeader(s.responseStatus)
		}

		if len(s.responseBody) != 0 {
			_, _ = w.Write(s.responseBody)
		}
	}))
}

func (s *ServerBuilder) JSON() *ServerBuilder {
	s.withJSON = true

	return s
}

func (s *ServerBuilder) OK() *ServerBuilder {
	s.responseStatus = http.StatusOK

	return s
}

func (s *ServerBuilder) Status(status int) *ServerBuilder {
	s.responseStatus = status

	return s
}

func (s *ServerBuilder) Body(body []byte) *ServerBuilder {
	s.responseBody = body

	return s
}

func (s *ServerBuilder) TextBody(body string) *ServerBuilder {
	return s.Body([]byte(body))
}
