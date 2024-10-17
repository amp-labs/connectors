package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
)

// Fixed is a server recipe that responds the same way regardless of the input.
type Fixed struct {
	// Setup is optional handler, where common http.ResponseWrite configuration takes place.
	Setup http.HandlerFunc
	// Always represents server handler that should implement how server should respond.
	Always http.HandlerFunc
}

// Server creates mock server.
func (f Fixed) Server() *httptest.Server {
	return Conditional{
		Setup: f.Setup,
		If: mockcond.Check(func(w http.ResponseWriter, r *http.Request) bool {
			return true
		}),
		Then: f.Always,
	}.Server()
}
