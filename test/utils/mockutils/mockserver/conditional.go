package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
)

// Conditional is a server recipe that describes how mock server should react when conditions are met.
// It is equivalent to If() Then{} Else{}.
type Conditional struct {
	// Setup is optional handler, where common http.ResponseWrite configuration takes place.
	Setup http.HandlerFunc
	// If, may consist of nested Or, And clauses allowing sophisticated logic.
	If mockcond.Condition
	// Then will be called when If evaluates to true.
	Then http.HandlerFunc
	// Else will be called when If evaluates to false.
	Else http.HandlerFunc
}

// Server creates mock server that will produce different response based on conditionals.
func (re Conditional) Server() *httptest.Server {
	// Reactive server is a simpler version of a Switch with one possible successful route.
	// This acts as syntactic sugar.
	return Switch{
		Setup: re.Setup,
		Cases: []Case{{
			If:   re.If,
			Then: re.Then,
		}},
		Default: re.Else,
	}.Server()
}
