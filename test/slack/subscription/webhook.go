package subscription

import (
	"encoding/json"
	"net/http"

	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func NewWebhookRouter() testscenario.WebhookRouter {
	return testscenario.WebhookRouter{
		Routes: []testscenario.WebhookRouteFunc{subscriptionConfirmation},
	}
}

// https://docs.slack.dev/reference/events/url_verification/
var subscriptionConfirmation = testscenario.WebhookRouteFunc(
	func(writer http.ResponseWriter, request *http.Request, data []byte) bool {
		body := requestBody{}
		if err := json.Unmarshal(data, &body); err != nil {
			return false
		}

		if body.Challenge == "" {
			return false
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "text/plain")

		// Bypassing HTML escaping is ok for this being used for local testing.
		_, _ = writer.Write([]byte(body.Challenge)) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter

		return true
	},
)

type requestBody struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}
