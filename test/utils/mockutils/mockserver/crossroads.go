package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
)

// Switch is a server recipe that describes multiple path a server may take.
// It is equivalent to Switch() {Case->Then...Case->Then} Default{}.
type Switch struct {
	// Setup is optional handler, where common http.ResponseWrite configuration takes place.
	Setup http.HandlerFunc
	// Cases is an ordered list of possible pathways a mock server could take.
	// The first case that satisfies a condition will be picked.
	Cases []Case
	// Default will be called only if all case had failed.
	Default http.HandlerFunc
}

// Server creates mock server that will produce different response based on conditionals.
func (c Switch) Server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Common setup is optional.
		if c.Setup != nil {
			c.Setup(w, r)
		}

		// Explore all paths stopping at the first satisfying server request requirement.
		for _, path := range c.Cases {
			if path.isOpen(w, r) {
				path.takeWalk(w, r)

				return
			}
		}

		// Server request has no matching resolution.
		if c.Default != nil {
			c.Default(w, r)

			return
		}

		// Default fail behaviour.
		w.WriteHeader(http.StatusInternalServerError)
		mockutils.WriteBody(w, `{"error": {"message": "condition failed"}}`)
	}))
}

// Case is one possible route a mock server can take if condition is satisfied.
type Case struct {
	// If, may consist of nested Or, And clauses allowing sophisticated logic.
	If mockcond.Condition
	// Then will be called when If evaluates to true.
	Then http.HandlerFunc
}

func (p *Case) isOpen(w http.ResponseWriter, r *http.Request) bool {
	if p.If == nil {
		return false
	}

	return p.If.EvaluateCondition(w, r)
}

func (p *Case) takeWalk(w http.ResponseWriter, r *http.Request) {
	if p.Then != nil {
		p.Then(w, r)

		return
	}

	// Default success behaviour.
	w.WriteHeader(http.StatusNoContent)

	return
}
