package smartlead

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCampaigns := testutils.DataFromFile(t, "read-campaign.json")
	responseCampaignsEmpty := testutils.DataFromFile(t, "read-campaign-empty.json")
	responseInvalidPath := testutils.DataFromFile(t, "read-invalid-path.html")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "email-accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Correct error message is understood from HTML response",
			Input: common.ReadParams{ObjectName: "email-accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidPath),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Cannot GET /api/v1/butterflies"),
			},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "email-accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Empty read response",
			Input: common.ReadParams{ObjectName: "campaigns", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseCampaignsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "campaigns", Fields: connectors.Fields("name", "status")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v1/campaigns"),
				Then:  mockserver.Response(http.StatusOK, responseCampaigns),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":   "Monthly sales",
						"status": "DRAFTED",
					},
					Raw: map[string]any{"id": float64(549768)},
				}, {
					Fields: map[string]any{
						"name":   "Weekly events",
						"status": "DRAFTED",
					},
					Raw: map[string]any{"id": float64(549772)},
				}, {
					Fields: map[string]any{
						"name":   "Black Friday prep",
						"status": "DRAFTED",
					},
					Raw: map[string]any{"id": float64(549773)},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
