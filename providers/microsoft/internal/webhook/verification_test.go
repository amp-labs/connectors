package webhook

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
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestVerifyWebhookMessage(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	_ = testutils.DataFromFile(t, "delete-not-found.json")

	tests := []testroutines.WebhookMessageVerification{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Write object and its ID must be included",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: nil,
				Params:  nil,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "Remove users",
			Input: testroutines.WebhookMessageVerificationParams{
				Request: nil,
				Params:  nil,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     true,
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (components.WebhookMessageVerifier, error) {
				return constructTestVerifier(tt.Server.URL)
			})
		})
	}
}

func constructTestVerifier(serverURL string) (*Verifier, error) {
	transport, err := components.NewTransport(providers.Microsoft, common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	transport.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(transport.ModuleInfo().BaseURL, serverURL))

	return NewVerifier(transport.JSONHTTPClient(), transport.ModuleInfo()), nil
}
