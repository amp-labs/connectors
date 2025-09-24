package linkedin

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the adAccounts",
			Input: common.WriteParams{ObjectName: "adAccounts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/adAccounts"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
					}),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, nil),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update adAccounts as POST",
			Input: common.WriteParams{
				ObjectName: "adAccounts",
				RecordId:   "514674276",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/adAccounts/514674276"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
						"X-Restli-Method":           []string{"PARTIAL_UPDATE"},
					}),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, nil),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: http.DefaultClient,
		Metadata: map[string]string{
			"adAccountId": "514674276",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(serverURL)

	return connector, nil
}
