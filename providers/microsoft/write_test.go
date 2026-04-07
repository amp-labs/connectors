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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseUser := testutils.DataFromFile(t, "write/users/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Create user via POST",
			Input: common.WriteParams{ObjectName: "users", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/users"),
				},
				Then: mockserver.Response(http.StatusOK, responseUser),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "520168be-3102-47bf-979d-7a99cb8ed8c5",
				Errors:   nil,
				Data: map[string]any{
					"@odata.context":    "https://graph.microsoft.com/v1.0/$metadata#users/$entity",
					"id":                "520168be-3102-47bf-979d-7a99cb8ed8c5",
					"businessPhones":    []any{},
					"displayName":       "Melissa Darrow",
					"givenName":         "Melissa",
					"jobTitle":          "Marketing Director",
					"mail":              nil,
					"mobilePhone":       "+1 206 555 0110",
					"officeLocation":    "131/1105",
					"preferredLanguage": "en-US",
					"surname":           "Darrow",
					"userPrincipalName": "MelissaD@integrationuserwithampersan.onmicrosoft.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update task via PATCH",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "520168be-3102-47bf-979d-7a99cb8ed8c5",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1.0/users/520168be-3102-47bf-979d-7a99cb8ed8c5"),
				},
				Then: mockserver.Response(http.StatusOK, responseUser),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "520168be-3102-47bf-979d-7a99cb8ed8c5",
				Errors:   nil,
				Data: map[string]any{
					"displayName": "Melissa Darrow",
					"givenName":   "Melissa",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create events vis POST",
			Input: common.WriteParams{ObjectName: "me/events", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/me/events"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"subject": "hello", "id": "753"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true, RecordId: "753", Data: map[string]any{"subject": "hello"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update events vis PATCH",
			Input: common.WriteParams{ObjectName: "me/events", RecordData: "dummy", RecordId: "723"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1.0/me/events/723"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"subject": "hello", "id": "753"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true, RecordId: "753", Data: map[string]any{"subject": "hello"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create calendars vis POST",
			Input: common.WriteParams{ObjectName: "calendars", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/me/calendars"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"subject": "hello", "id": "753"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true, RecordId: "753", Data: map[string]any{"subject": "hello"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update calendars vis PATCH",
			Input: common.WriteParams{ObjectName: "calendars", RecordData: "dummy", RecordId: "723"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1.0/me/calendars/723"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"subject": "hello", "id": "753"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true, RecordId: "753", Data: map[string]any{"subject": "hello"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create messages vis POST",
			Input: common.WriteParams{ObjectName: "me/messages", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1.0/me/messages"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"subject": "hello", "id": "753"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true, RecordId: "753", Data: map[string]any{"subject": "hello"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update messages vis PATCH",
			Input: common.WriteParams{ObjectName: "me/messages", RecordData: "dummy", RecordId: "723"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v1.0/me/messages/723"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"subject": "hello", "id": "753"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true, RecordId: "753", Data: map[string]any{"subject": "hello"},
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
