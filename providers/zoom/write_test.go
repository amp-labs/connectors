package zoom

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

func TestWriteModuleMeeting(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseCreateTrackingField := testutils.DataFromFile(t, "./write/create-tracking-fields.json")
	responseUpdateTrackingField := testutils.DataFromFile(t, "./write/create-tracking-fields.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "calendarList"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "aero", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},

		{
			Name:  "Create tracking fields as POST",
			Input: common.WriteParams{ObjectName: "tracking_fields", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateTrackingField),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "a32CJji-weJ92",
				Errors:   nil,
				Data: map[string]any{
					"id":       "a32CJji-weJ92",
					"field":    "field1",
					"required": false,
					"visible":  true,
				},
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Update tracking fields as PUT",
			Input: common.WriteParams{ObjectName: "tracking_fields", RecordId: "a32CJji-weJ92", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, responseUpdateTrackingField),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "a32CJji-weJ92",
				Errors:   nil,
				Data: map[string]any{
					"id":       "a32CJji-weJ92",
					"field":    "field1",
					"required": false,
					"visible":  true,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleMeeting)
			})
		})
	}
}

func TestWriteModuleUser(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseCreateUser := testutils.DataFromFile(t, "./write/create-users.json")
	responseCreateContactsGroup := testutils.DataFromFile(t, "./write/create-contacts-groups.json")

	tests := []testroutines.Write{
		{
			Name:  "Create user as POST",
			Input: common.WriteParams{ObjectName: "users", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateUser),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "KDcuGIm1QgePTO8WbOqwIQ",
				Errors:   nil,
				Data: map[string]any{
					"id":         "KDcuGIm1QgePTO8WbOqwIQ",
					"email":      "jchill@example.com",
					"first_name": "Jill",
					"last_name":  "Chill",
				},
			},
			ExpectedErrs: nil,
		},

		{
			Name:  "Create contact group as POST",
			Input: common.WriteParams{ObjectName: "contacts_groups", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responseCreateContactsGroup),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "A4ql1FjgL913r",
				Errors:   nil,
				Data: map[string]any{
					"group_id":    "A4ql1FjgL913r",
					"group_name":  "Developers",
					"description": "A contact group.",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleUser)
			})
		})
	}
}
