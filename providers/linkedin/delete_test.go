package linkedin

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestAdsDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorNotFound := testutils.DataFromFile(t, "delete-missing-adAccounts.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "adAccounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "adAccounts", RecordId: "514674276"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/adAccounts/514674276"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
					}),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Error on deleting missing record",
			Input: common.DeleteParams{ObjectName: "adAccounts", RecordId: "516445454"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError(
					"Not Found.",
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestAdsConnector(tt.Server.URL)
			})
		})
	}
}

func TestPlatformDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "posts", RecordId: "urn:li:share:7393604235420078080"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/posts/urn:li:share:7393604235420078080"),
					mockcond.Header(http.Header{
						"LinkedIn-Version":          []string{"202504"},
						"X-Restli-Protocol-Version": []string{"2.0.0"},
					}),
					mockcond.MethodDELETE(),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestPlatformConnector(tt.Server.URL)
			})
		})
	}
}
