package google

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorPageToken := testutils.DataFromFile(t, "read/error-page-token.json")
	errorNotFound := testutils.DataFromFile(t, "read/error-object-not-found.html")
	responseCalendarListFirstPage := testutils.DataFromFile(t, "read/calendarList/1-first-page.json")
	responseCalendarListLastPage := testutils.DataFromFile(t, "read/calendarList/2-last-page.json")
	responseSettingsFirstPage := testutils.DataFromFile(t, "read/settings/1-first-page.json")
	responseSettingsLastPage := testutils.DataFromFile(t, "read/settings/2-last-page.json")
	responseEventsFirstPage := testutils.DataFromFile(t, "read/events/1-first-page.json")
	responseEventsLastPage := testutils.DataFromFile(t, "read/events/2-last-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "calendarList"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "calendarList", Fields: connectors.Fields("summary")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorPageToken),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Invalid page token value."), // nolint:goerr113
			},
		},
		{
			Name:  "Error endpoint for object is not found",
			Input: common.ReadParams{ObjectName: "calendarList", Fields: connectors.Fields("summary")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				errors.New("The requested URL /calendar/v3/calendarList?maxResults=3000&showDeleted=true was not found on this server."), // nolint:goerr113,lll
			},
		},
		{
			Name: "Read calendarList first page",
			Input: common.ReadParams{
				ObjectName: "calendarList",
				Fields:     connectors.Fields("summary", "description"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/calendar/v3/users/me/calendarList"),
				Then:  mockserver.Response(http.StatusOK, responseCalendarListFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 4,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"summary":     "Holidays in Tanzania",
						"description": "Holidays and Observances in Tanzania",
					},
					Raw: map[string]any{
						"id": "en.tz#holiday@group.v.calendar.google.com",
					},
				}, {
					Fields: map[string]any{
						"summary":     "Holidays in United States",
						"description": "Holidays and Observances in United States",
					},
					Raw: map[string]any{
						"id": "en.usa#holiday@group.v.calendar.google.com",
					},
				}, {
					Fields: map[string]any{
						"summary":     "Holidays in Ukraine",
						"description": "Holidays and Observances in Ukraine",
					},
					Raw: map[string]any{
						"id": "en.ukrainian#holiday@group.v.calendar.google.com",
					},
				}, {
					Fields: map[string]any{
						"summary":     "Holidays in India",
						"description": "Holidays and Observances in India",
					},
					Raw: map[string]any{
						"id": "en.indian#holiday@group.v.calendar.google.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/calendar/v3/users/me/calendarList?maxResults=3000" +
					"&pageToken=EjAKDAjR0u-8BhCAjJaYAxIgdTpnY2FsK2dyb3VwOi8vaG9saWRheS9lbi5pbmRpYW4aHhIJBwYs8u00rJIAqsmIjAQNEgsIk9TvvAYQgM7SWQ==", // nolint:lll
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read calendarList second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "calendarList",
				Fields:     connectors.Fields("summary", "description"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/calendar/v3/users/me/calendarList"),
					mockcond.QueryParam("maxResults", "3000"),
				},
				Then: mockserver.Response(http.StatusOK, responseCalendarListLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"summary": "integration@test.com",
					},
					Raw: map[string]any{
						"id": "integration@test.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read settings first page",
			Input: common.ReadParams{
				ObjectName: "settings",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/calendar/v3/users/me/settings"),
				Then:  mockserver.Response(http.StatusOK, responseSettingsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 10,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "autoAddHangouts",
					},
					Raw: map[string]any{
						"value": "false",
					},
				}},
				NextPage: testroutines.URLTestServer + "/calendar/v3/users/me/settings?maxResults=3000" +
					"&pageToken=CiEKCwoJL2NhbHVzZXIvEgQIAhIAGgxzaG93RGVjbGluZWQQoJLNg62eiwM=", // nolint:lll
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read settings second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "settings",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/calendar/v3/users/me/settings"),
				Then:  mockserver.Response(http.StatusOK, responseSettingsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "timezone",
					},
					Raw: map[string]any{
						"value": "America/Los_Angeles",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read events first page using incremental approach",
			Input: common.ReadParams{
				ObjectName: "events",
				Fields:     connectors.Fields("id"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/calendar/v3/calendars/primary/events"),
					mockcond.QueryParam("updatedMin", "2024-09-19T12:30:45.000Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseEventsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "4of57j0lu6oqjdo62om7juoe3k",
					},
					Raw: map[string]any{
						"summary": "[Meeting Buffer]",
						"colorId": "8",
						"iCalUID": "4of57j0lu6oqjdo62om7juoe3k@google.com",
					},
				}, {
					Fields: map[string]any{
						"id": "cl392f4vinvbes6dnv2r0j8ku0",
					},
					Raw: map[string]any{
						"summary":     "Meeting with Integration",
						"hangoutLink": "https://meet.google.com/puo-ojnc-wir",
						"iCalUID":     "cl392f4vinvbes6dnv2r0j8ku0@google.com",
					},
				}, {
					Fields: map[string]any{
						"id": "md15kq7kv4p1viuta3jip2inng",
					},
					Raw: map[string]any{
						"created": "2025-06-02T09:43:16.000Z",
						"updated": "2025-06-12T18:58:09.447Z",
						"summary": "Meeting with Integration",
					},
				}},
				NextPage: testroutines.URLTestServer + "/calendar/v3/calendars/primary/events?maxResults=3000" +
					"&updatedMin=2024-09-19T12:30:45.000Z" +
					"&pageToken=CkAKMAouCgwIwcaswgYQmOig1QESHgocChptZDE1a3E3a3Y0cDF2aXV0YTNqaXAyaW5uZxoMCNfVwMMGENC33MkBwD4B", // nolint:lll
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read events second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "events",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/calendar/v3/calendars/primary/events"),
				Then:  mockserver.Response(http.StatusOK, responseEventsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "p3nkh1hg41683vdhlcq4k12iso",
					},
					Raw: map[string]any{
						"created": "2025-06-02T09:43:48.000Z",
						"updated": "2025-06-12T18:58:09.447Z",
						"summary": "Meeting with Integration",
						"iCalUID": "p3nkh1hg41683vdhlcq4k12iso@google.com",
					},
				}},
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
				return constructTestCalendarConnector(tt.Server.URL)
			})
		})
	}
}
