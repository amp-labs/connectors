package microsoft

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	errorUnknownResource := testutils.DataFromFile(t, "read/unknown-resource.json")
	responseUsersFirst := testutils.DataFromFile(t, "read/users/1-first-page.json")
	responseUsersLast := testutils.DataFromFile(t, "read/users/2-second-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorUnknownResource),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, testutils.StringError("Resource not found for the segment 'user'."),
			},
		},
		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("displayName")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1.0/users"),
				Then:  mockserver.Response(http.StatusOK, responseUsersFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"displayname": "Integration User",
					},
					Raw: map[string]any{
						"surname":           "User",
						"userPrincipalName": "integration.user_withampersand.com#EXT#@integrationuserwithampersan.onmicrosoft.com",
						"id":                "12151ea6-6d86-4afd-a68d-88ab34f5170a",
					},
				}},
				NextPage: "https://graph.microsoft.com/v1.0/users?$top=1&$skiptoken=RFNwdAIAAQAAACM6aW50ZWdyYXRpb24udXNlckB3aXRoYW1wZXJzYW5kLmNvbSlVc2VyXzEyMTUxZWE2LTZkODYtNGFmZC1hNjhkLTg4YWIzNGY1MTcwYbkAAAAAAAAAAAAA", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("displayName"),
				NextPage:   testroutines.URLTestServer + "/v1.0/users?$top=1&$skiptoken=RFNwdAIAAQAAACM6aW50ZWdyYXRpb24udXNlckB3aXRoYW1wZXJzYW5kLmNvbSlVc2VyXzEyMTUxZWE2LTZkODYtNGFmZC1hNjhkLTg4YWIzNGY1MTcwYbkAAAAAAAAAAAAA", // nolint:lll
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1.0/users"),
				Then:  mockserver.Response(http.StatusOK, responseUsersLast),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
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
