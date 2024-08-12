package smartlead

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
			Name:         "Mime response header expected",
			Input:        common.ReadParams{ObjectName: "contact"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Correct error message is understood from HTML response",
			Input: common.ReadParams{ObjectName: "contact"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseInvalidPath)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Cannot GET /api/v1/butterflies"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "contact"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{}`)
			})),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Empty read response",
			Input: common.ReadParams{ObjectName: "campaigns"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseCampaignsEmpty)
			})),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "campaigns", Fields: []string{"name", "status"}},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseCampaigns)
			})),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				// custom comparison focuses on subset of fields to keep the test short
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.NextPage.String() == expected.NextPage.String() &&
					actual.Done == expected.Done
			},
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
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
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
