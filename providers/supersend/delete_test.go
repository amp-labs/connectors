package supersend

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
			Input:        common.DeleteParams{ObjectName: "labels"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Unsupported object returns error",
			Input:        common.DeleteParams{ObjectName: "unsupported", RecordId: "123"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Delete senders fails (no delete endpoint)",
			Input:        common.DeleteParams{ObjectName: "senders", RecordId: "sender-001"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Delete teams fails (no delete endpoint)",
			Input:        common.DeleteParams{ObjectName: "teams", RecordId: "team-001"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Labels - soft delete",
			Input: common.DeleteParams{ObjectName: "labels", RecordId: "label-456"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v1/labels/label-456"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Campaigns - delete",
			Input: common.DeleteParams{ObjectName: "campaigns", RecordId: "campaign-001"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v1/auto/campaign/campaign-001"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Contacts - delete",
			Input: common.DeleteParams{ObjectName: "contacts", RecordId: "contact-001"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v2/contacts/contact-001"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Sender Profiles - soft delete",
			Input: common.DeleteParams{ObjectName: "sender-profiles", RecordId: "sp-001"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/v1/sender-profile/sp-001"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
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
