package manage

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) {
	t.Parallel()

	responseActivitiesEmpty := testutils.DataFromFile(t, "read-activities-empty.json")
	responseActivitiesSinglePage := testutils.DataFromFile(t, "read-activities-single-page.json")
	firstPageFixture := testutils.DataFromFile(t, "read-activities-first-page.json")
	responseLaukCivil := testutils.DataFromFile(t,
		"read-lauk-civil-controlled.json")

	tests := []testroutines.Read{
		{
			Name: "Read activities empty list",
			Input: common.ReadParams{
				ObjectName: "activities",
				Fields:     connectors.Fields("id", "etag"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v4/activities.json"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
				},
				Then: mockserver.Response(http.StatusOK, responseActivitiesEmpty),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "read first page",
			Input: common.ReadParams{
				ObjectName: "activities",
				Fields:     connectors.Fields("id", "etag"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v4/activities.json"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
				},
				Then: mockserver.Response(http.StatusOK, firstPageFixture),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(125565935),
						"etag": "\"77bfbef69f3e1596a68f6d0d2c363349\"",
					},
					Raw: map[string]any{
						"id":   float64(125565935),
						"etag": "\"77bfbef69f3e1596a68f6d0d2c363349\"",
					},
				}},
				NextPage: "https://eu.app.clio.com/api/v4/activities.json?limit=1&order=id%28asc%29&page_token=xyx",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read activities with updated_since",
			Input: common.ReadParams{
				ObjectName: "activities",
				Fields:     connectors.Fields("id", "etag"),
				PageSize:   1,
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v4/activities.json"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
					mockcond.QueryParam("updated_since", "2026-04-01T00:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseActivitiesSinglePage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(125565936),
						"etag": "\"88cfcf07af4e2696b79f7e1e3d474450\"",
					},
					Raw: map[string]any{
						"id":   float64(125565936),
						"etag": "\"88cfcf07af4e2696b79f7e1e3d474450\"",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read lauk_civil_controlled_rates with Since omits updated_since and filters by updated_at client-side",
			Input: common.ReadParams{
				ObjectName: "lauk_civil_controlled_rates",
				Fields:     connectors.Fields("id", "etag", "updated_at"),
				PageSize:   50,
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v4/lauk_civil_controlled_rates.json"),
					mockcond.QueryParam("limit", "50"),
					mockcond.QueryParam("order", "id(asc)"),
					mockcond.QueryParamsMissing("updated_since"),
				},
				Then: mockserver.Response(http.StatusOK, responseLaukCivil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"etag":       "e2",
						"updated_at": "2026-05-01T00:00:00Z",
					},
					Raw: map[string]any{
						"etag":       "e2",
						"updated_at": "2026-05-01T00:00:00Z",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read lauk_civil_controlled_rates without Since returns all rows and omits updated_since",
			Input: common.ReadParams{
				ObjectName: "lauk_civil_controlled_rates",
				Fields:     connectors.Fields("id", "etag"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v4/lauk_civil_controlled_rates.json"),
					mockcond.QueryParam("limit", "50"),
					mockcond.QueryParam("order", "id(asc)"),
					mockcond.QueryParamsMissing("updated_since"),
				},
				Then: mockserver.Response(http.StatusOK, responseLaukCivil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":   float64(1),
							"etag": "e1",
						},
						Raw: map[string]any{
							"id":         float64(1),
							"etag":       "e1",
							"updated_at": "2026-03-01T00:00:00Z",
						},
					},
					{
						Fields: map[string]any{
							"id":   float64(2),
							"etag": "e2",
						},
						Raw: map[string]any{
							"id":         float64(2),
							"etag":       "e2",
							"updated_at": "2026-05-01T00:00:00Z",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read activities next page via NextPage",
			Input: common.ReadParams{
				ObjectName: "activities",
				Fields:     connectors.Fields("id", "etag"),
				PageSize:   1,
				NextPage:   testroutines.URLTestServer + "/api/v4/activities.json?limit=1&order=id(asc)&page_token=xyz", //nolint:lll
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v4/activities.json"),
					mockcond.QueryParam("limit", "1"),
					mockcond.QueryParam("order", "id(asc)"),
					mockcond.QueryParam("page_token", "xyz"),
				},
				Then: mockserver.Response(http.StatusOK, responseActivitiesSinglePage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   float64(125565936),
						"etag": "\"88cfcf07af4e2696b79f7e1e3d474450\"",
					},
					Raw: map[string]any{
						"id":   float64(125565936),
						"etag": "\"88cfcf07af4e2696b79f7e1e3d474450\"",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructReadTestAdapter(tt.Server.URL)
			})
		})
	}
}

func constructReadTestAdapter(serverURL string) (*Adapter, error) {
	adapter, err := NewAdapter(common.ConnectorParams{
		Module:              providers.ModuleClioManage,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "app.clio.com",
		Metadata: map[string]string{
			"region": "",
		},
	})
	if err != nil {
		return nil, err
	}

	adapter.SetBaseURL(mockutils.ReplaceURLOrigin(adapter.HTTPClient().Base, serverURL))

	return adapter, nil
}
