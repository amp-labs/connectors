// nolint:dupl
package linkedin

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestAdsWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
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
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("X-Restli-Id", "514674276"),
					mockserver.Response(http.StatusOK, nil),
				),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "514674276",
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
				return constructTestAdsConnector(tt.Server.URL)
			})
		})
	}
}

func TestPlatformWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the posts",
			Input: common.WriteParams{ObjectName: "posts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/posts"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
					}),
					mockcond.MethodPOST(),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("X-Restli-Id", "urn:li:share:7393604235420078080"),
					mockserver.Response(http.StatusOK, nil),
				),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "urn:li:share:7393604235420078080",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update posts as POST",
			Input: common.WriteParams{
				ObjectName: "posts",
				RecordId:   "urn:li:share:7393604235420078080",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/posts/urn:li:share:7393604235420078080"),
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
				return constructTestPlatformConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestPlatformConnector(serverURL string) (*Connector, error) {
	return constructTestConnector(serverURL, providers.ModuleLinkedInPlatform, nil)
}

func constructTestAdsConnector(serverURL string) (*Connector, error) {
	return constructTestConnector(serverURL, providers.ModuleLinkedInAds, map[string]string{"adAccountId": "514674276"})
}

func constructTestConnector(serverURL string, moduleID common.ModuleID, metadata map[string]string) (*Connector, error) { //nolint:lll
	connector, err := NewConnector(common.ConnectorParams{
		Module:              moduleID,
		AuthenticatedClient: http.DefaultClient,
		Metadata:            metadata,
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
