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

func TestCalendarRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorPageToken := testutils.DataFromFile(t, "calendar/read/error-page-token.json")
	errorNotFound := testutils.DataFromFile(t, "calendar/read/error-object-not-found.html")
	responseCalendarListFirstPage := testutils.DataFromFile(t, "calendar/read/calendarList/1-first-page.json")
	responseCalendarListLastPage := testutils.DataFromFile(t, "calendar/read/calendarList/2-last-page.json")
	responseSettingsFirstPage := testutils.DataFromFile(t, "calendar/read/settings/1-first-page.json")
	responseSettingsLastPage := testutils.DataFromFile(t, "calendar/read/settings/2-last-page.json")
	responseEventsFirstPage := testutils.DataFromFile(t, "calendar/read/events/1-first-page.json")
	responseEventsLastPage := testutils.DataFromFile(t, "calendar/read/events/2-last-page.json")

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

func TestContactsRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorPageSize := testutils.DataFromFile(t, "contacts/read/error-bad-page-size.json")
	errorNotFound := testutils.DataFromFile(t, "contacts/read/error-object-not-found.html")
	responseContactGroupsFirstPage := testutils.DataFromFile(t, "contacts/read/contactGroups/1-first-page.json")
	responseContactGroupsLastPage := testutils.DataFromFile(t, "contacts/read/contactGroups/2-last-page.json")
	responseMyConnectionsFirstPage := testutils.DataFromFile(t, "contacts/read/myConnections/1-first-page.json")
	responseMyConnectionsLastPage := testutils.DataFromFile(t, "contacts/read/myConnections/2-last-page.json")
	responseOtherContacts := testutils.DataFromFile(t, "contacts/read/otherContacts/one-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contactGroups"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "contactGroups", Fields: connectors.Fields("memberCount")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorPageSize),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Page size must be less than or equal to 1000."), // nolint:goerr113
			},
		},
		{
			Name:  "Error endpoint for object is not found",
			Input: common.ReadParams{ObjectName: "contactGroups", Fields: connectors.Fields("memberCount")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				errors.New("The requested URL /v1/bananas was not found on this server."), // nolint:goerr113
			},
		},
		{
			Name: "Read contactGroups first page",
			Input: common.ReadParams{
				ObjectName: "contactGroups",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/contactGroups"),
				Then:  mockserver.Response(http.StatusOK, responseContactGroupsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "starred",
						"name": "starred",
					},
					Raw: map[string]any{
						"resourceName":  "contactGroups/starred",
						"groupType":     "SYSTEM_CONTACT_GROUP",
						"formattedName": "Starred",
					},
				}, {
					Fields: map[string]any{
						"id":   "friends",
						"name": "friends",
					},
					Raw: map[string]any{
						"resourceName":  "contactGroups/friends",
						"groupType":     "SYSTEM_CONTACT_GROUP",
						"formattedName": "Friends",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v1/contactGroups?groupFields=name&pageSize=1000" +
					"&pageToken=CAISDAjEt9vDBhDoq6i8Ag",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contactGroups second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "contactGroups",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/contactGroups"),
					mockcond.QueryParam("pageSize", "1000"),
				},
				Then: mockserver.Response(http.StatusOK, responseContactGroupsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "blocked",
						"name": "blocked",
					},
					Raw: map[string]any{
						"resourceName":  "contactGroups/blocked",
						"groupType":     "SYSTEM_CONTACT_GROUP",
						"formattedName": "Blocked",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read myConnections first page",
			Input: common.ReadParams{
				ObjectName: "myConnections",
				Fields:     connectors.Fields("id", "resourceName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/people/me/connections"),
				Then:  mockserver.Response(http.StatusOK, responseMyConnectionsFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":           "c3194283054244587340",
						"resourcename": "people/c3194283054244587340",
					},
					Raw: map[string]any{
						"etag": "%EgUBAi43PRoEAQIFByIMZHUyb2I4ZWlGdVk9",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v1/people/me/connections?pageSize=1000" +
					"&pageToken=GiAKHAgBagsIvd7bwwYQiLuICXILCLTe28MGEOChvEYQAg",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read myConnections second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "myConnections",
				Fields:     connectors.Fields("id", "resourceName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/people/me/connections"),
				Then:  mockserver.Response(http.StatusOK, responseMyConnectionsLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":           "c3614676454703510757",
						"resourcename": "people/c3614676454703510757",
					},
					Raw: map[string]any{
						"etag": "%EgUBAi43PRoEAQIFByIMZWthZ1VlTWN4eEE9",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read otherContacts one and only page",
			Input: common.ReadParams{
				ObjectName: "otherContacts",
				Fields:     connectors.Fields("id", "resourceName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/otherContacts"),
				Then:  mockserver.Response(http.StatusOK, responseOtherContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":           "c6949365612481516512",
						"resourcename": "otherContacts/c6949365612481516512",
					},
					Raw: map[string]any{
						"etag": "%EgcBAgkuNz0+GgQBAgUHIgxJVnA4dkJSUTBQMD0=",
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
				return constructTestContactsConnector(tt.Server.URL)
			})
		})
	}
}
