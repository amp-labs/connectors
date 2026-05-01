package subscription

import (
	"net/http"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func NewWebhookRouter() testscenario.WebhookRouter {
	return testscenario.WebhookRouter{
		Routes: []testscenario.Route{subscriptionConfirmation},
	}
}

// Default handling.
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#decrypting-resource-data-from-change-notifications
var subscriptionConfirmation = testscenario.Route{
	// This route is executed when Microsoft is verifying that webhook is rechable.
	Left: func(request *http.Request) bool {
		url, err := urlbuilder.FromRawURL(request.URL)
		if err != nil {
			return false
		}

		return url.HasQueryParam("validationToken")
	},
	// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#notificationurl-validation
	Right: func(writer http.ResponseWriter, request *http.Request) {
		url, err := urlbuilder.FromRawURL(request.URL)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		validationToken, ok := url.GetFirstQueryParam("validationToken")
		if !ok {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "text/plain")
		_, _ = writer.Write([]byte(validationToken))
	},
}
