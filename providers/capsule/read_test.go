package capsule

import (
	"errors"
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

	responseNotFoundError := testutils.DataFromFile(t, "read/not-found.json")
	responsePartiesFirstPage := testutils.DataFromFile(t, "read/parties/first-page.json")
	responsePartiesLastPage := testutils.DataFromFile(t, "read/parties/last-page.json")
	responseProjects := testutils.DataFromFile(t, "read/projects/first-page.json")

	tests := []testroutines.Read{
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "categories", Fields: connectors.Fields("colour")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFoundError),
			}.Server(),
			ExpectedErrs: []error{
				errors.New("Could not find resource"),
				common.ErrBadRequest,
			},
		},
		{
			Name: "Read parties first page incrementally",
			Input: common.ReadParams{
				ObjectName: "parties",
				Fields:     connectors.Fields("id", "type"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/parties"),
					// Pacific time to UTC is achieved by adding 8 hours
					mockcond.QueryParam("since", "2024-09-19T12:30:45Z"),
					mockcond.QueryParam("embed", "fields"),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://api.capsulecrm.com/api/v2/parties?page=2&perPage=2>; rel="next"`),
					mockserver.Response(http.StatusOK, responsePartiesFirstPage),
				),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(254633973),
						"type": "organisation",
					},
					Raw: map[string]any{
						"name": "Capsule",
					},
				}, {
					Fields: map[string]any{
						"id":   float64(254633972),
						"type": "person",
					},
					Raw: map[string]any{
						"lastName": "User",
					},
				}},
				NextPage: "https://api.capsulecrm.com/api/v2/parties?page=2&perPage=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read parties second page which is without next cursor",
			Input: common.ReadParams{
				ObjectName: "parties",
				Fields:     connectors.Fields("id", "type"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/parties"),
				Then:  mockserver.Response(http.StatusOK, responsePartiesLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(267944762),
						"type": "person",
					},
					Raw: map[string]any{
						"firstName": "integration.tes1907",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read projects which have custom fields",
			Input: common.ReadParams{
				ObjectName: "projects",
				Fields:     connectors.Fields("id", "name", "interests"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/kases"),
					mockcond.QueryParam("embed", "fields"),
				},
				Then: mockserver.Response(http.StatusOK, responseProjects),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(5588202),
						"name": "Research",
						// Custom fields can be requested.
						"interests": "Skiing",
					},
					Raw: map[string]any{
						"description": "Designing shopping cart website",
						// Custom fields are at the root level
						"fields": []any{
							map[string]any{
								"id": float64(9785121),
								"definition": map[string]any{
									"id":   float64(926886),
									"name": "Interests",
								},
								"value": "Skiing",
								"tagId": float64(168298),
							},
						},
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
