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

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	organizationIdResponse := testutils.DataFromFile(t, "organization-id.json")
	membersResponse := testutils.DataFromFile(t, "members.json")
	rolesResponse := testutils.DataFromFile(t, "roles.json")
	billingcentersResponse := testutils.DataFromFile(t, "billingcenters.json")
	tests := []testroutines.Read{
		{
			Name:  "Read list of members",
			Input: common.ReadParams{ObjectName: "members", Fields: connectors.Fields("")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v1/organizations/5cf59a25-5063-40e1-826b-5ceaf369b207/members"),
					Then: mockserver.Response(http.StatusOK, membersResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"sub_request_status": "SUCCESS",
							"member": map[string]any{
								"id":                "82ebdbce-de88-4ba0-a6b4-e77d51a2ce99",
								"email":             "integration.user+snapchat@withampersand.com",
								"organization_id":   "5cf59a25-5063-40e1-826b-5ceaf369b207",
								"display_name":      "Integration User",
								"snapchat_username": "ampersandonset",
								"member_status":     "MEMBER",
							},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of roles",
			Input: common.ReadParams{ObjectName: "roles", Fields: connectors.Fields("")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v1/organizations/5cf59a25-5063-40e1-826b-5ceaf369b207/roles"),
					Then: mockserver.Response(http.StatusOK, rolesResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"sub_request_status": "SUCCESS",
							"role": map[string]any{
								"id":              "548a683e-181b-4021-ad61-4e1964ab3111",
								"container_kind":  "Organizations",
								"container_id":    "5cf59a25-5063-40e1-826b-5ceaf369b207",
								"member_id":       "82ebdbce-de88-4ba0-a6b4-e77d51a2ce99",
								"organization_id": "5cf59a25-5063-40e1-826b-5ceaf369b207",
								"type":            "admin",
							},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of billingcenters",
			Input: common.ReadParams{ObjectName: "billingcenters", Fields: connectors.Fields("")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v1/organizations/5cf59a25-5063-40e1-826b-5ceaf369b207/billingcenters"),
					Then: mockserver.Response(http.StatusOK, billingcentersResponse),
				}, {
					If:   mockcond.Path("/v1/me"),
					Then: mockserver.Response(http.StatusOK, organizationIdResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"sub_request_status": "SUCCESS",
							"billingcenter": map[string]any{
								"id":              "002291c6-878b-4c9e-b690-6c0fccab3dce",
								"organization_id": "5cf59a25-5063-40e1-826b-5ceaf369b207",
								"name":            "New Billing Center",
								"email_address":   "honeybear_Itd@example. com",
							},
						},
					},
				},
				NextPage: "https://adsapi.snapchat.com/v1/organizations/5cf59a25-5063-40e1-826b-5ceaf369b207/billingcenters" +
					"?cursor=QmlsbGluZ0NlbnRlcnMvMDAyMjkxYzYtODc4Yi00YzllLWI2OTAtNmMwZmNjYWIzZGNl&limit=100",
				Done: false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
