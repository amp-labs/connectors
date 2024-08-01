package mockserver

import (
	"net/http"
	"net/http/httptest"

	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver/mockresponse"
)

// Select will create a server that will have multiple choices for response.
// It will go in order and try to see the first choice that satisfies connectors request.
// Once found it will use the respective response.
func Select(choices ...*mockresponse.Plan) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, plan := range choices {
			// Try to match if this request matches response plan.
			if plan.Check(r) {
				// We found how to respond.
				plan.OnSuccess(w, r)
				return
			}
		}

		// If nothing matched we didn't respond appropriately.
		// Mock server should break the invalid test.
		w.WriteHeader(http.StatusInternalServerError)
	}))
}
