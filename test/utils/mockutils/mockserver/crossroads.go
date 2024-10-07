package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
)

// Crossroad is a server recipe that describes multiple path a server may take based on .
type Crossroad struct {
	// Setup is optional handler, where common http.ResponseWrite configuration takes place.
	Setup http.HandlerFunc
	// Paths is an ordered list of possible pathways a mock server could take.
	// The first path that satisfies a condition will be picked.
	Paths []Path
	// OnFailure will be called only if all paths are blocked.
	OnFailure http.HandlerFunc
}

// Server creates mock server that will produce different response based on conditionals.
func (c Crossroad) Server() *httptest.Server {
	return NewServer(func(w http.ResponseWriter, r *http.Request) {
		// Common setup is optional.
		if c.Setup != nil {
			c.Setup(w, r)
		}

		// Explore all paths stopping at the first satisfying server request requirement.
		for _, path := range c.Paths {
			if path.isOpen(w, r) {
				path.takeWalk(w, r)

				return
			}
		}

		// Server request has no matching resolution.
		if c.OnFailure != nil {
			c.OnFailure(w, r)

			return
		}

		// Default fail behaviour.
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": {"message": "condition failed"}}`))
	})
}

// Path is one possible route a mock server can take if condition is satisfied.
type Path struct {
	// Condition may consist of nested Or, And clauses allowing sophisticated logic.
	Condition mockcond.Condition
	// OnSuccess will be called if Condition evaluates to true.
	OnSuccess http.HandlerFunc
}

func (p *Path) isOpen(w http.ResponseWriter, r *http.Request) bool {
	if p.Condition == nil {
		return false
	}

	return p.Condition.RunCondition(w, r)
}

func (p *Path) takeWalk(w http.ResponseWriter, r *http.Request) {
	if p.OnSuccess != nil {
		p.OnSuccess(w, r)

		return
	}

	// Default success behaviour.
	w.WriteHeader(http.StatusNoContent)

	return
}
