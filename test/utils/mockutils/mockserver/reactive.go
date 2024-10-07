package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils"
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
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Common setup is optional.
		if re.Setup != nil {
			re.Setup(w, r)
		}

		re.apply(w, r)
	}))
}

// Apply will try to resolve conditions and find respective scenario to execute.
// If check condition is satisfied then OnSuccess is executed, otherwise OnFailure.
// There is a default behaviour for each leaf case.
func (re Reactive) apply(w http.ResponseWriter, r *http.Request) {
	if re.Condition.RunCondition(w, r) {
		if re.OnSuccess != nil {
			re.OnSuccess(w, r)

			return
		}

		// Default success behaviour.
		w.WriteHeader(http.StatusNoContent)

		return
	}

	if re.OnFailure != nil {
		re.OnFailure(w, r)

		return
	}

	// Default fail behaviour.
	w.WriteHeader(http.StatusInternalServerError)
	mockutils.WriteBody(w, `{"error": {"message": "condition failed"}}`)
}
