package snapchatads

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	organizationIdResponse := testutils.DataFromFile(t, "organization-id.json")
	billingcentersResponse := testutils.DataFromFile(t, "write_billingcenters.json")
	rolesResponse := testutils.DataFromFile(t, "write_roles.json")

	tests := []testroutines.Write{
		{
			Name:  "Create a billingcenters as POST",
			Input: common.WriteParams{ObjectName: "billingcenters", RecordData: "dummy"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.MethodPOST(),
					Then: mockserver.Response(http.StatusOK, billingcentersResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6e0f4532-3702-4f0b-9889-9fe5d0614afd",
				Errors:   nil,
				Data: map[string]any{
					"request_status": "SUCCESS",
					"request_id":     "5eaa9f3f00ff0e03450355dbc60001737e616473617069",
					"billingcenters": []any{
						map[string]any{
							"sub_request_status": "SUCCESS",
							"billingcenter": map[string]any{
								"id":                              "6e0f4532-3702-4f0b-9889-9fe5d0614afd",
								"updated_at":                      "2025-02-30T09:49:52.118Z",
								"created_at":                      "2025-02-30T09:49:52.118Z",
								"organization_id":                 "8fdeefec-f502-4ca8-9a84-6411e0a51053",
								"name":                            "Kianjous Billing Center",
								"email_address":                   "honeybear_ltd@example.com",
								"address_line_1":                  "10 Honey Bear Road",
								"locality":                        "London",
								"administrative_district_level_1": "GB-LND",
								"country":                         "GB",
								"postal_code":                     "NW1 4RY",
								"alternative_email_addresses":     []any{"honeybear_burrow@example.com"},
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update billingcenters as PUT",
			Input: common.WriteParams{
				ObjectName: "billingcenters",
				RecordId:   "6e0f4532-3702-4f0b-9889-9fe5d0614afd",
				RecordData: "dummy",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.MethodPUT(),
					Then: mockserver.Response(http.StatusOK, billingcentersResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6e0f4532-3702-4f0b-9889-9fe5d0614afd",
				Errors:   nil,
				Data: map[string]any{
					"request_status": "SUCCESS",
					"request_id":     "5eaa9f3f00ff0e03450355dbc60001737e616473617069",
					"billingcenters": []any{
						map[string]any{
							"sub_request_status": "SUCCESS",
							"billingcenter": map[string]any{
								"id":                              "6e0f4532-3702-4f0b-9889-9fe5d0614afd",
								"updated_at":                      "2025-02-30T09:49:52.118Z",
								"created_at":                      "2025-02-30T09:49:52.118Z",
								"organization_id":                 "8fdeefec-f502-4ca8-9a84-6411e0a51053",
								"name":                            "Kianjous Billing Center",
								"email_address":                   "honeybear_ltd@example.com",
								"address_line_1":                  "10 Honey Bear Road",
								"locality":                        "London",
								"administrative_district_level_1": "GB-LND",
								"country":                         "GB",
								"postal_code":                     "NW1 4RY",
								"alternative_email_addresses":     []any{"honeybear_burrow@example.com"},
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create roles as POST",
			Input: common.WriteParams{ObjectName: "roles", RecordData: "dummy"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.MethodPOST(),
					Then: mockserver.Response(http.StatusOK, rolesResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9a121178-73f0-4089-8f80-a60e5445e8ea",
				Errors:   nil,
				Data: map[string]any{
					"request_status": "SUCCESS",
					"request_id":     "5eaa9f3f00ff0e03450355dbc60001737e616473617069",
					"roles": []any{
						map[string]any{
							"sub_request_status": "SUCCESS",
							"role": map[string]any{
								"id":              "9a121178-73f0-4089-8f80-a60e5445e8ea",
								"updated_at":      "2020-04-28T13:44:37.586Z",
								"created_at":      "2020-04-28T13:44:37.586Z",
								"container_kind":  "Organizations",
								"container_id":    "8fdeefec-f502-4ca8-9a84-6411e0a51053",
								"member_id":       "d051973d-32b2-496b-b44a-345986bce17d",
								"organization_id": "8fdeefec-f502-4ca8-9a84-6411e0a51053",
								"type":            "member",
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update roles as PUT",
			Input: common.WriteParams{
				ObjectName: "roles",
				RecordData: "dummy",
				RecordId:   "9a121178-73f0-4089-8f80-a60e5445e8ea",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.MethodPUT(),
					Then: mockserver.Response(http.StatusOK, rolesResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9a121178-73f0-4089-8f80-a60e5445e8ea",
				Errors:   nil,
				Data: map[string]any{
					"request_status": "SUCCESS",
					"request_id":     "5eaa9f3f00ff0e03450355dbc60001737e616473617069",
					"roles": []any{
						map[string]any{
							"sub_request_status": "SUCCESS",
							"role": map[string]any{
								"id":              "9a121178-73f0-4089-8f80-a60e5445e8ea",
								"updated_at":      "2020-04-28T13:44:37.586Z",
								"created_at":      "2020-04-28T13:44:37.586Z",
								"container_kind":  "Organizations",
								"container_id":    "8fdeefec-f502-4ca8-9a84-6411e0a51053",
								"member_id":       "d051973d-32b2-496b-b44a-345986bce17d",
								"organization_id": "8fdeefec-f502-4ca8-9a84-6411e0a51053",
								"type":            "member",
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

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
