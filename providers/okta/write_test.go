package okta

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	createUserResponse := testutils.DataFromFile(t, "write-user-create.json")
	updateUserResponse := testutils.DataFromFile(t, "write-user-update.json")
	createGroupResponse := testutils.DataFromFile(t, "write-group-create.json")
	updateGroupResponse := testutils.DataFromFile(t, "write-group-update.json")
	errorResponse := testutils.DataFromFile(t, "error.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Create user successfully",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordData: map[string]any{
					"profile": map[string]any{
						"firstName": "Test",
						"lastName":  "User",
						"email":     "test.user@example.com",
						"login":     "test.user@example.com",
					},
				},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, createUserResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "00u_newuser_12345",
				Data: map[string]any{
					"id":     "00u_newuser_12345",
					"status": "STAGED",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update user successfully",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "00u1234567890abcdef",
				RecordData: map[string]any{
					"profile": map[string]any{
						"firstName": "Updated",
						"lastName":  "Name",
					},
				},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, updateUserResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "00u1234567890abcdef",
				Data: map[string]any{
					"id":     "00u1234567890abcdef",
					"status": "ACTIVE",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create group successfully",
			Input: common.WriteParams{
				ObjectName: "groups",
				RecordData: map[string]any{
					"profile": map[string]any{
						"name":        "New Group",
						"description": "A newly created group",
					},
				},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, createGroupResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "00g_newgroup_12345",
				Data: map[string]any{
					"id":   "00g_newgroup_12345",
					"type": "OKTA_GROUP",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update group successfully (uses PUT method)",
			Input: common.WriteParams{
				ObjectName: "groups",
				RecordId:   "00g1234567890abcdef",
				RecordData: map[string]any{
					"profile": map[string]any{
						"name":        "Updated Group Name",
						"description": "Updated group description",
					},
				},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, updateGroupResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "00g1234567890abcdef",
				Data: map[string]any{
					"id":   "00g1234567890abcdef",
					"type": "OKTA_GROUP",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Write returns error on bad request",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordData: map[string]any{},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
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
