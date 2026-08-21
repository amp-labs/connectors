package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
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
	responseCompaniesWithAssociations := testutils.DataFromFile(t, "companies_with_associations.json")

	tests := []testconn.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
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
				If: mockcond.And{
					mockcond.Path("/v2/objects/companies/records/query"),
					// Attio expects a singular "filter" key with record_id $in.
					mockcond.Body(`{
						"filter": {
							"record_id": {
								"$in": ["1bdb55e3-67f4-48d3-829b-45db3039a960", "3a95b53c-e7a1-4e53-a4e4-436f72283818"]
							}
						}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseGetRecordsByIds),
			}.Server(),

			Expected: []common.ReadResultRow{
				{
					Id: "1bdb55e3-67f4-48d3-829b-45db3039a960",
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
					Id: "3a95b53c-e7a1-4e53-a4e4-436f72283818",
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
		{
			Name: "Fetch companies by IDs with record-reference associations",
			Input: GetRecordsByIdsInput{
				ObjectName:   "companies",
				Fields:       []string{"name"},
				Ids:          []string{"2db97cee-6c6b-4486-ae52-db8e4b6f44e9"},
				Associations: []string{"people"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/objects/companies/records/query"),
					mockcond.Body(`{
						"filter": {
							"record_id": {
								"$in": ["2db97cee-6c6b-4486-ae52-db8e4b6f44e9"]
							}
						}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCompaniesWithAssociations),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "2db97cee-6c6b-4486-ae52-db8e4b6f44e9",
					Fields: map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2025-03-12T07:55:38.981000000Z",
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
					Associations: map[string][]common.Association{
						"people": {
							{ObjectId: "891dcbfc-9141-415d-9b2a-2238a6cc012d"},
							{ObjectId: "5e3fb280-007b-495a-a530-9354bde01de1"},
						},
					},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "63d34516-b287-4c27-9d28-fe2adbebcd50",
							"object_id":    "1fa986a6-952e-4e92-ba01-acca61a7b616",
							"record_id":    "2db97cee-6c6b-4486-ae52-db8e4b6f44e9",
						},
						"created_at": "2025-03-12T07:55:38.327000000Z",
						"values": map[string]any{
							"name": []any{
								map[string]any{
									"active_from":  "2025-03-12T07:55:38.981000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Attio",
									"attribute_type": "text",
								},
							},
							"team": []any{
								map[string]any{
									"active_from":      "2025-03-12T07:55:39.000000000Z",
									"active_until":     nil,
									"created_by_actor": map[string]any{"type": "workspace-member", "id": "073f4c74-b60d-4de9-992a-0f799b5e442e"},
									"attribute_type":   "record-reference",
									"target_object":    "people",
									"target_record_id": "891dcbfc-9141-415d-9b2a-2238a6cc012d",
								},
								map[string]any{
									"active_from":      "2025-03-12T07:55:39.100000000Z",
									"active_until":     nil,
									"created_by_actor": map[string]any{"type": "workspace-member", "id": "073f4c74-b60d-4de9-992a-0f799b5e442e"},
									"attribute_type":   "record-reference",
									"target_object":    "people",
									"target_record_id": "5e3fb280-007b-495a-a530-9354bde01de1",
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

			result, err := conn.GetRecordsByIds(
				t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, tt.Input.Associations,
			)

			tt.Validate(t, err, result)
		})
	}
}
