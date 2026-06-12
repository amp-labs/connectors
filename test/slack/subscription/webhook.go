package subscription

import (
	"net/http"

	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func NewWebhookRouter() testscenario.WebhookRouter {
	return testscenario.WebhookRouter{
		Routes: []testscenario.Route{subscriptionConfirmation},
	}
}

// Default handling.
// https://docs.slack.dev/reference/events/url_verification/
var subscriptionConfirmation = testscenario.Route{
	// This route is executed when Microsoft is verifying that webhook is rechable.
	Left: func(request *http.Request) bool {
		body, err := httpkit.ReadJSONBody[requestBody](request)
		if err != nil {
			return false
		}

		return body.Challenge != ""
	},
	Right: func(writer http.ResponseWriter, request *http.Request) {
		body, err := httpkit.ReadJSONBody[requestBody](request)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "text/plain")

		// Bypassing HTML escaping is ok for this being used for local testing.
		_, _ = writer.Write([]byte(body.Challenge)) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
	},
}

type requestBody struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}
