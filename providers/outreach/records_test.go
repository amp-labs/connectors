package outreach

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	accountsByIdsResponse := testutils.DataFromFile(t, "accounts_by_ids.json")

	tests := []testroutines.TestCase[common.ReadByIdsParams, []common.ReadResultRow]{
		{
			Name:         "Missing object name returns error",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully fetch accounts by IDs",
			Input: common.ReadByIdsParams{
				ObjectName: "accounts",
				Fields:     []string{"name"},
				RecordIds:  []string{"1", "2", "5"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/accounts"),
					mockcond.QueryParam("filter[id]", "1,2,5"),
				},
				Then: mockserver.Response(http.StatusOK, accountsByIdsResponse),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "1",
					Fields: map[string]any{
						"id":   float64(1),
						"name": "Example Corp 1",
					},
					// Raw preserves the original JSON structure: "attributes" key is NOT removed.
					// Fields are extracted from the flattened record (attributes merged to top level).
					Raw: map[string]any{
						"id":   float64(1),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-15T10:30:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "First test account",
							"domain":            "example1.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Technology",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "San Francisco",
							"name":              "Example Corp 1",
							"named":             true,
							"naturalName":       "Example Corp 1",
							"numberOfEmployees": float64(100),
							"tags":              []any{},
							"touchedAt":         "2025-01-15T10:30:00.000Z",
							"updatedAt":         "2025-01-15T10:30:00.000Z",
							"websiteUrl":        "https://example1.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(1),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/1",
						},
					},
				},
				{
					Id: "2",
					Fields: map[string]any{
						"id":   float64(2),
						"name": "Example Corp 2",
					},
					Raw: map[string]any{
						"id":   float64(2),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-16T11:45:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "Second test account",
							"domain":            "example2.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Finance",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "New York",
							"name":              "Example Corp 2",
							"named":             true,
							"naturalName":       "Example Corp 2",
							"numberOfEmployees": float64(250),
							"tags":              []any{},
							"touchedAt":         "2025-01-16T11:45:00.000Z",
							"updatedAt":         "2025-01-16T11:45:00.000Z",
							"websiteUrl":        "https://example2.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(2),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/2",
						},
					},
				},
				{
					Id: "5",
					Fields: map[string]any{
						"id":   float64(5),
						"name": "Example Corp 5",
					},
					Raw: map[string]any{
						"id":   float64(5),
						"type": "account",
						"attributes": map[string]any{
							"buyerIntentScore":  nil,
							"companyType":       nil,
							"createdAt":         "2025-01-17T09:20:00.000Z",
							"custom1":           nil,
							"customId":          nil,
							"description":       "Third test account",
							"domain":            "example5.com",
							"externalSource":    "outreach-api",
							"followers":         float64(0),
							"foundedAt":         nil,
							"industry":          "Healthcare",
							"linkedInEmployees": nil,
							"linkedInUrl":       nil,
							"locality":          "Boston",
							"name":              "Example Corp 5",
							"named":             true,
							"naturalName":       "Example Corp 5",
							"numberOfEmployees": float64(500),
							"tags":              []any{},
							"touchedAt":         "2025-01-17T09:20:00.000Z",
							"updatedAt":         "2025-01-17T09:20:00.000Z",
							"websiteUrl":        "https://example5.com",
						},
						"relationships": map[string]any{
							"owner": map[string]any{
								"data": map[string]any{
									"type": "user",
									"id":   float64(1),
								},
							},
						},
						"links": map[string]any{
							"self": "https://api.outreach.io/api/v2/accounts/5",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input)

			tt.Validate(t, err, result)
		})
	}
}
