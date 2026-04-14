package subscriber

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestCreate(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testroutines.CreateSubscription{
		{
			Name: "Missing request data",
			Input: common.SubscribeParams{
				Request: nil,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Invalid request type",
			Input: common.SubscribeParams{
				Request: "invalid",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errInvalidRequestType},
		},
		{
			Name: "Successful subscription",
			Input: common.SubscribeParams{
				Request: &Input{
					ChangeType:         "created",
					WebhookURL:         "https://example.com/webhook",
					Resource:           "me/messages",
					ExpirationDateTime: "2026-04-13T21:15:00Z",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then: mockserver.ResponseString(http.StatusOK, `{
					"id": "sub-id",
					"changeType": "created",
					"notificationUrl": "https://example.com/webhook",
					"resource": "me/messages",
					"expirationDateTime": "2026-04-13T21:15:00Z"
				}`),
			}.Server(),
			Expected: &common.SubscriptionResult{
				Result: &Output{
					ChangeType:         "created",
					WebhookURL:         "https://example.com/webhook",
					Resource:           "me/messages",
					ExpirationDateTime: "2026-04-13T21:15:00Z",
				},
				Status: common.SubscriptionStatusSuccess,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (components.SubscriptionCreator, error) {
				return constructTestStrategy(tt.Server.URL)
			})
		})
	}
}

func constructTestStrategy(serverURL string) (*Strategy, error) {
	transport, err := components.NewTransport(providers.Microsoft, common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(transport.ModuleInfo().BaseURL, serverURL))

	return NewStrategy(transport.JSONHTTPClient(), transport.ModuleInfo(), nil), nil
}
