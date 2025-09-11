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

func TestCalendarWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorCalendarNoColor := testutils.DataFromFile(t, "calendar/write/calendarList/error-missing-color.json")
	responseInsertCalendar := testutils.DataFromFile(t, "calendar/write/calendarList/new.json")
	responseEvent := testutils.DataFromFile(t, "calendar/write/events/new.json")

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
				errors.New(
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

func TestContactsWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseContactGroups := testutils.DataFromFile(t, "contacts/write/contactGroups/new.json")
	responseMyConnections := testutils.DataFromFile(t, "contacts/write/myConnections/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "contactGroups"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Create contact group",
			Input: common.WriteParams{
				ObjectName: "contactGroups",
				RecordData: map[string]any{"data": "value"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/contactGroups"),
					mockcond.Body(`{"contactGroup": {"data": "value"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactGroups),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "50ea5f3188536092",
				Errors:   nil,
				Data: map[string]any{
					"resourceName":  "contactGroups/50ea5f3188536092",
					"etag":          "EDU2I4Eaqmg=",
					"formattedName": "Virgie Wisozk",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact group",
			Input: common.WriteParams{
				ObjectName: "contactGroups",
				RecordData: map[string]any{"data": "value"},
				RecordId:   "50ea5f3188536092",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/v1/contactGroups/50ea5f3188536092"),
					mockcond.Body(`{"contactGroup": {"data": "value"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactGroups),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "50ea5f3188536092",
				Errors:   nil,
				Data: map[string]any{
					"resourceName":  "contactGroups/50ea5f3188536092",
					"etag":          "EDU2I4Eaqmg=",
					"formattedName": "Virgie Wisozk",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create my connections",
			Input: common.WriteParams{
				ObjectName: "myConnections",
				RecordData: map[string]any{"names": "value", "etag": "value2"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.QueryParamsMissing("updatePersonFields"),
					mockcond.Path("/v1/people:createContact"),
					mockcond.Body(`{"names": "value", "etag": "value2"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseMyConnections),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "c2600607527630254940",
				Errors:   nil,
				Data: map[string]any{
					"resourceName": "people/c2600607527630254940",
					"etag":         "%EigBAgMEBQYHCAkKCwwNDg8QERITFBUWFxkfISIjJCUmJy40NTc9Pj9AGgQBAgUHIgxSUjBKd1IyLzdoaz0=",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update my connections",
			Input: common.WriteParams{
				ObjectName: "myConnections",
				RecordData: map[string]any{"names": "value", "etag": "value2"},
				RecordId:   "c2600607527630254940",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.QueryParam("updatePersonFields", "names"), // etag field is omitted
					mockcond.Path("/v1/people/c2600607527630254940:updateContact"),
					mockcond.Body(`{"names": "value", "etag": "value2"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseMyConnections),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "c2600607527630254940",
				Errors:   nil,
				Data: map[string]any{
					"resourceName": "people/c2600607527630254940",
					"etag":         "%EigBAgMEBQYHCAkKCwwNDg8QERITFBUWFxkfISIjJCUmJy40NTc9Pj9AGgQBAgUHIgxSUjBKd1IyLzdoaz0=",
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
				return constructTestContactsConnector(tt.Server.URL)
			})
		})
	}
}

func TestMailWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseDrafts := testutils.DataFromFile(t, "mail/write/drafts/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "drafts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Create contact group",
			Input: common.WriteParams{ObjectName: "drafts", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/gmail/v1/users/me/drafts"),
				},
				Then: mockserver.Response(http.StatusOK, responseDrafts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "r5151461288000502025",
				Errors:   nil,
				Data: map[string]any{
					"id": "r5151461288000502025",
					"message": map[string]any{
						"id":       "199400f21a8f1186",
						"threadId": "199400f21a8f1186",
						"labelIds": []any{
							"DRAFT",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update contact group",
			Input: common.WriteParams{
				ObjectName: "drafts",
				RecordData: "dummy",
				RecordId:   "r5151461288000502025",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/gmail/v1/users/me/drafts/r5151461288000502025"),
				},
				Then: mockserver.Response(http.StatusOK, responseDrafts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "r5151461288000502025",
				Errors:   nil,
				Data: map[string]any{
					"id": "r5151461288000502025",
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
				return constructTestMailConnector(tt.Server.URL)
			})
		})
	}
}
