package gong

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

type GetRecordsByIdsInput struct {
	ObjectName string
	Ids        []string
	Fields     []string
}

// nolint:lll,funlen
func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	callsResp := testutils.DataFromFile(t, "read.json")
	usersResp := testutils.DataFromFile(t, "get-records-users.json")

	tests := []testroutines.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name: "Calls by IDs POST to /v2/calls/extensive with callIds filter and contentSelector",
			Input: GetRecordsByIdsInput{
				ObjectName: "calls",
				Ids:        []string{"52947912500572621", "137982752092261989"},
				Fields:     []string{"id"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/calls/extensive"),
					mockcond.Body(`{
						"filter":{"callIds":["52947912500572621","137982752092261989"]},
						"contentSelector":{"context":"Extended","exposedFields":{"parties":true,"media":true}}
					}`),
				},
				Then: mockserver.Response(http.StatusOK, callsResp),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id:     "52947912500572621",
					Fields: map[string]any{"id": "52947912500572621"},
					Raw: map[string]any{
						"metaData": map[string]any{
							"id":             "52947912500572621",
							"url":            "https://us-49467.app.gong.io/call?id=52947912500572621",
							"workspaceId":    "1007648505208900737",
							"clientUniqueId": "ce93bb26-de69-41e3-8a7f-43ea3714b9e8",
							"customData":     "R1201",
						},
					},
				},
				{
					Id:     "137982752092261989",
					Fields: map[string]any{"id": "137982752092261989"},
					Raw: map[string]any{
						"metaData": map[string]any{
							"id":             "137982752092261989",
							"url":            "https://us-49467.app.gong.io/call?id=137982752092261989",
							"workspaceId":    "1007648505208900737",
							"clientUniqueId": "f77501df-0c70-4c38-b565-a3a09fee14fb",
							"customData":     "R1201",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Users by IDs POST to /v2/users/extensive with userIds filter and no contentSelector",
			Input: GetRecordsByIdsInput{
				ObjectName: "users",
				Ids:        []string{"8000000000000001", "8000000000000002"},
				Fields:     []string{"id", "emailAddress", "firstName", "lastName", "active", "title"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/users/extensive"),
					mockcond.Body(`{"filter":{"userIds":["8000000000000001","8000000000000002"]}}`),
				},
				Then: mockserver.Response(http.StatusOK, usersResp),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "8000000000000001",
					Fields: map[string]any{
						"id":           "8000000000000001",
						"emailaddress": "alice@example.com",
						"firstname":    "Alice",
						"lastname":     "Anderson",
						"active":       true,
						"title":        "Account Executive",
					},
					Raw: map[string]any{
						"id":           "8000000000000001",
						"emailAddress": "alice@example.com",
						"firstName":    "Alice",
						"lastName":     "Anderson",
						"active":       true,
						"title":        "Account Executive",
					},
				},
				{
					Id: "8000000000000002",
					Fields: map[string]any{
						"id":           "8000000000000002",
						"emailaddress": "bob@example.com",
						"firstname":    "Bob",
						"lastname":     "Brown",
						"active":       true,
						"title":        "Sales Engineer",
					},
					Raw: map[string]any{
						"id":           "8000000000000002",
						"emailAddress": "bob@example.com",
						"firstName":    "Bob",
						"lastName":     "Brown",
						"active":       true,
						"title":        "Sales Engineer",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Fields filter limits Fields map to requested fields (plus id) while Raw is untouched",
			Input: GetRecordsByIdsInput{
				ObjectName: "users",
				Ids:        []string{"8000000000000001", "8000000000000002"},
				Fields:     []string{"emailAddress"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/users/extensive"),
					mockcond.Body(`{"filter":{"userIds":["8000000000000001","8000000000000002"]}}`),
				},
				Then: mockserver.Response(http.StatusOK, usersResp),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "8000000000000001",
					Fields: map[string]any{
						"id":           "8000000000000001",
						"emailaddress": "alice@example.com",
					},
					Raw: map[string]any{
						"id":           "8000000000000001",
						"emailAddress": "alice@example.com",
						"firstName":    "Alice",
						"lastName":     "Anderson",
						"active":       true,
						"title":        "Account Executive",
					},
				},
				{
					Id: "8000000000000002",
					Fields: map[string]any{
						"id":           "8000000000000002",
						"emailaddress": "bob@example.com",
					},
					Raw: map[string]any{
						"id":           "8000000000000002",
						"emailAddress": "bob@example.com",
						"firstName":    "Bob",
						"lastName":     "Brown",
						"active":       true,
						"title":        "Sales Engineer",
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			t.Cleanup(tt.Close)

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(t.Context(),
				tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, nil)

			tt.Validate(t, err, result)
		})
	}
}
