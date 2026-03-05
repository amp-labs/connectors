package greenhouse

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseError := testutils.DataFromFile(t, "error.json")
	responseCandidatesFirstPage := testutils.DataFromFile(t, "read/candidates/first-page.json")
	responseCandidatesLastPage := testutils.DataFromFile(t, "read/candidates/last-page.json")

	tests := []testroutines.Read{
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "candidates", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Your request included invalid JSON."),
			},
		},
		{
			Name: "Read candidates first page with pagination",
			Input: common.ReadParams{
				ObjectName: "candidates",
				Fields:     connectors.Fields("id", "first_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3/candidates"),
					mockcond.QueryParam("per_page", "500"),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://harvest.greenhouse.io/v3/candidates?cursor=abc123>; rel="next"`),
					mockserver.Response(http.StatusOK, responseCandidatesFirstPage),
				),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         float64(12345),
						"first_name": "John",
					},
					Raw: map[string]any{
						"last_name": "Doe",
					},
				}, {
					Fields: map[string]any{
						"id":         float64(67890),
						"first_name": "Jane",
					},
					Raw: map[string]any{
						"last_name": "Smith",
					},
				}},
				NextPage: "https://harvest.greenhouse.io/v3/candidates?cursor=abc123",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read candidates last page without next link",
			Input: common.ReadParams{
				ObjectName: "candidates",
				Fields:     connectors.Fields("id", "first_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/candidates"),
				Then:  mockserver.Response(http.StatusOK, responseCandidatesLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         float64(11111),
						"first_name": "Alice",
					},
					Raw: map[string]any{
						"last_name": "Johnson",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read candidates incrementally with updated_after",
			Input: common.ReadParams{
				ObjectName: "candidates",
				Fields:     connectors.Fields("id", "first_name"),
				Since: time.Date(2024, 9, 1, 10, 0, 0, 0,
					time.FixedZone("UTC-5", -5*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3/candidates"),
					mockcond.QueryParam("per_page", "500"),
					mockcond.QueryParam("updated_after", "2024-09-01T15:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseCandidatesLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         float64(11111),
						"first_name": "Alice",
					},
					Raw: map[string]any{
						"last_name": "Johnson",
					},
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
