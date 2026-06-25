package subscription

import (
	"net/http"

	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

func NewWebhookRouter() testscenario.WebhookRouter {
	return testscenario.WebhookRouter{
		Routes: []testscenario.WebhookRouteFunc{subscriptionConfirmation},
	}
}

// Default handling.
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#decrypting-resource-data-from-change-notifications
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#notificationurl-validation
var subscriptionConfirmation = testscenario.WebhookRouteFunc(
	func(writer http.ResponseWriter, request *http.Request, data []byte) bool {
		url, err := urlbuilder.FromRawURL(request.URL)
		// During the Subscription creation Microsoft contacts webhook to verify that it is rechable.
		if err != nil {
			return false
		}

		if !url.HasQueryParam("validationToken") {
			// Not a verification request.
			return false
		}

		validationToken, ok := url.GetFirstQueryParam("validationToken")
		if !ok {
			writer.WriteHeader(http.StatusInternalServerError)
			return true
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "text/plain")
		_, _ = writer.Write([]byte(validationToken))

		return true
	},
)
