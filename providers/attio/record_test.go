package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// GetRecordsByIdsInput represents the input parameters for GetRecordsByIds method.
type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

func TestGetRecordByIds(t *testing.T) {
	t.Parallel()

	responseGetRecordsByIds := testutils.DataFromFile(t, "get_records_by_ids.json")

	tests := []testroutines.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name:         "Missing object name returns error",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully fetch companies by IDs",
			Input: GetRecordsByIdsInput{
				ObjectName: "companies",
				Fields:     []string{"name"},
				Ids:        []string{"1bdb55e3-67f4-48d3-829b-45db3039a960", "3a95b53c-e7a1-4e53-a4e4-436f72283818"},
			},

			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/objects/companies/records/query"),
				Then:  mockserver.Response(http.StatusOK, responseGetRecordsByIds),
			}.Server(),

			Expected: []common.ReadResultRow{
				{
					Fields: map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2026-01-29T08:22:24.888000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "system",
									"id":   nil,
								},
								"value":          "Apple",
								"attribute_type": "text",
							},
						},
					},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "e8d74639-96e5-41be-af46-ced812aef5c5",
							"object_id":    "8381cb5a-fc74-4421-aaad-3092c8bea210",
							"record_id":    "1bdb55e3-67f4-48d3-829b-45db3039a960",
						},
						"created_at": "2026-01-29T08:22:24.253000000Z",
						"web_url":    "https://app.attio.com/ampersand-test/company/1bdb55e3-67f4-48d3-829b-45db3039a960",
						"values": map[string]any{
							"record_id": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:24.253000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "1bdb55e3-67f4-48d3-829b-45db3039a960",
									"attribute_type": "text",
								},
							},
							"name": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:24.888000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Apple",
									"attribute_type": "text",
								},
							},
						},
					},
				},
				{
					Fields: map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2026-01-29T08:22:23.198000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "system",
									"id":   nil,
								},
								"value":          "Attio",
								"attribute_type": "text",
							},
						},
					},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "e8d74639-96e5-41be-af46-ced812aef5c5",
							"object_id":    "8381cb5a-fc74-4421-aaad-3092c8bea210",
							"record_id":    "3a95b53c-e7a1-4e53-a4e4-436f72283818",
						},
						"created_at": "2026-01-29T08:22:22.656000000Z",
						"web_url":    "https://app.attio.com/ampersand-test/company/3a95b53c-e7a1-4e53-a4e4-436f72283818",
						"values": map[string]any{
							"record_id": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:22.656000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "3a95b53c-e7a1-4e53-a4e4-436f72283818",
									"attribute_type": "text",
								},
							},
							"name": []any{
								map[string]any{
									"active_from":  "2026-01-29T08:22:23.198000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Attio",
									"attribute_type": "text",
								},
							},
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

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, nil)

			tt.Validate(t, err, result)
		})
	}
}
