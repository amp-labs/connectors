package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
)

// Conditional is a server recipe that describes how mock server should react when conditions are met.
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
func (c Conditional) Server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Common setup is optional.
		if c.Setup != nil {
			c.Setup(w, r)
		}

		c.apply(w, r)
	}))
}

// Apply will try to resolve conditions and find respective scenario to execute.
// If check condition is satisfied Then is executed, otherwise Else.
// There is a default behaviour for each leaf case.
func (c Conditional) apply(w http.ResponseWriter, r *http.Request) {
	if c.If.EvaluateCondition(w, r) {
		if c.Then != nil {
			c.Then(w, r)

			return
		}

		// Default success behaviour.
		w.WriteHeader(http.StatusNoContent)

		return
	}

	if c.Else != nil {
		c.Else(w, r)

		return
	}

	// Default fail behaviour.
	w.WriteHeader(http.StatusInternalServerError)
	mockutils.WriteBody(w, `{"error": {"message": "condition failed"}}`)
}
