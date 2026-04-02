package okta

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Delete user successfully",
			Input: common.DeleteParams{ObjectName: "users", RecordId: "00u1234567890abcdef"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete group successfully",
			Input: common.DeleteParams{ObjectName: "groups", RecordId: "00g1234567890abcdef"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete trustedOrigin successfully",
			Input: common.DeleteParams{ObjectName: "trustedOrigins", RecordId: "tos1234567890abcdef"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete zone successfully",
			Input: common.DeleteParams{ObjectName: "zones", RecordId: "nzo1234567890abcdef"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete eventHook successfully",
			Input: common.DeleteParams{ObjectName: "eventHooks", RecordId: "who1234567890abcdef"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Delete authorizationServer successfully",
			Input: common.DeleteParams{ObjectName: "authorizationServers", RecordId: "aus1234567890abcdef"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
