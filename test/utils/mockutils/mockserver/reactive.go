package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
)

// Reactive is a server recipe that describes how mock server should react when conditions are met.
type Reactive struct {
	// Setup is optional handler, where common http.ResponseWrite configuration takes place.
	Setup http.HandlerFunc
	// Condition may consist of nested Or, And clauses allowing sophisticated logic.
	Condition mockcond.Condition
	// OnSuccess will be called if Condition evaluates to true.
	OnSuccess http.HandlerFunc
	// OnFailure will be called if Condition evaluates to false.
	OnFailure http.HandlerFunc
}

// Server creates mock server that will produce different response based on conditionals.
func (re Reactive) Server() *httptest.Server {
	// Reactive server is a simpler version of a Crossroad with one possible successful route.
	// This acts as syntactic sugar.
	return Crossroad{
		Setup: re.Setup,
		Paths: []Path{{
			Condition: re.Condition,
			OnSuccess: re.OnSuccess,
		}},
		OnFailure: re.OnFailure,
	}.Server()
}
