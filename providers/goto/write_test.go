package gotoconn

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read-only object rejects write",
			Input: common.WriteParams{
				ObjectName: "sessions",
				RecordData: map[string]any{"foo": "bar"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create webinar via POST returns webinarKey",
			Input: common.WriteParams{
				ObjectName: "webinars",
				RecordData: map[string]any{
					"subject":     "Intro",
					"description": "Hello",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/G2W/rest/v2/organizers/" + testAccountKey + "/webinars"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"webinarKey":"7878787878787878"}`)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "7878787878787878",
				Data: map[string]any{
					"webinarKey": "7878787878787878",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update webinar via PATCH preserves caller-supplied id",
			Input: common.WriteParams{
				ObjectName: "webinars",
				RecordId:   "7878787878787878",
				RecordData: map[string]any{"subject": "Renamed"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/G2W/rest/v2/organizers/" + testAccountKey + "/webinars/7878787878787878"),
				},
				Then: mockserver.Response(http.StatusNoContent, nil),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "7878787878787878",
				Data:     map[string]any{},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "SCIM update uses PUT, not PATCH",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "user-123",
				RecordData: map[string]any{"displayName": "Alice"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/identity/v1/Users/user-123"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"id":"user-123","displayName":"Alice"}`)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "user-123",
				Data: map[string]any{
					"id":          "user-123",
					"displayName": "Alice",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
