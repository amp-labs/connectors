package google

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorCalendarNoColor := testutils.DataFromFile(t, "write/calendarList/error-missing-color.json")
	responseInsertCalendar := testutils.DataFromFile(t, "write/calendarList/new.json")
	responseEvent := testutils.DataFromFile(t, "write/events/new.json")

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
			Name:  "Bad request from provider",
			Input: common.WriteParams{ObjectName: "calendarList", RecordData: make(map[string]any)},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorCalendarNoColor),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Missing foreground color.",
				),
			},
		},
		{
			Name: "Valid insert of a calendar item to my calendar",
			Input: common.WriteParams{
				ObjectName: "calendarList",
				RecordData: map[string]any{
					"foregroundColor": "#ffffff",
					"backgroundColor": "#19f7f0",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/calendar/v3/users/me/calendarList"),
					mockcond.QueryParam("colorRgbFormat", "true"),
				},
				Then: mockserver.Response(http.StatusOK, responseInsertCalendar),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "en.italian#holiday@group.v.calendar.google.com",
				Errors:   nil,
				Data: map[string]any{
					"summary":     "Holidays in Italy",
					"description": "Holidays and Observances in Italy",
					"timeZone":    "America/Los_Angeles",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Write must act as an Update",
			Input: common.WriteParams{
				ObjectName: "calendarList",
				RecordId:   "en.usa#holiday@group.v.calendar.google.com",
				RecordData: make(map[string]any),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/calendar/v3/users/me/calendarList/en.usa#holiday@group.v.calendar.google.com"), // nolint:lll
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Create event",
			Input: common.WriteParams{
				ObjectName: "events",
				RecordData: map[string]any{},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/calendar/v3/calendars/primary/events"),
				},
				Then: mockserver.Response(http.StatusOK, responseEvent),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "std4ien2l4fovmbup80ov20tp8",
				Errors:   nil,
				Data: map[string]any{
					"status":  "confirmed",
					"summary": "Monthly team meeting",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestCalendarConnector(tt.Server.URL)
			})
		})
	}
}
