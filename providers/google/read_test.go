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
				errors.New("Invalid page token value."),
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
				errors.New("The requested URL /calendar/v3/calendarList?maxResults=3000&showDeleted=true was not found on this server."), // nolint:lll
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
				errors.New("Page size must be less than or equal to 1000."),
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
				errors.New("The requested URL /v1/bananas was not found on this server."),
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

func TestMailRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorForbidden := testutils.DataFromFile(t, "mail/forbidden.json")
	errorNotFound := testutils.DataFromFile(t, "mail/not-found.html")
	responseMessagesFirstPage := testutils.DataFromFile(t, "mail/read/messages/1-first-page.json")
	responseMessagesLastPage := testutils.DataFromFile(t, "mail/read/messages/2-last-page.json")
	responseMessageItem := testutils.DataFromFile(t, "mail/read/messages/message-item.json")
	responseDrafts := testutils.DataFromFile(t, "mail/read/drafts/drafts.json")
	responseDraftMessageItem1 := testutils.DataFromFile(t, "mail/read/drafts/message-item-1.json")
	responseDraftMessageItem2 := testutils.DataFromFile(t, "mail/read/drafts/message-item-2.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "messages"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error forbidden object",
			Input: common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, errorForbidden),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden,
				errors.New("CSE is not enabled."), // nolint:goerr113
			},
		},
		{
			Name:  "HTML Error for not found object",
			Input: common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				common.ErrNotFound,
				errors.New("The requested URL /gmail/v1/users/me/butterfly was not found on this server. That’s all we know"), // nolint:goerr113,lll
			},
		},
		{
			Name: "Read messages first page",
			Input: common.ReadParams{
				ObjectName: "messages",
				Fields:     connectors.Fields("id"),
				PageSize:   33,
				Since: time.Date(2024, 9, 19, 23, 0, 0, 0,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/gmail/v1/users/me/messages"),
					mockcond.QueryParam("maxResults", "33"),      // from params
					mockcond.QueryParam("q", "after:2024/09/20"), // it is 20 due to time zone
				},
				Then: mockserver.Response(http.StatusOK, responseMessagesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "1993fb4b539b5a1a",
					},
					Raw: map[string]any{
						"threadId": "1993fa50bd191f7b",
					},
				}, {
					Fields: map[string]any{
						"id": "1993fa50bd191f7b",
					},
					Raw: map[string]any{
						"threadId": "1993fa50bd191f7b",
					},
				}},
				NextPage: "08277485409175924556",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read messages second page without next cursor",
			Input: common.ReadParams{
				ObjectName: "messages",
				Fields:     connectors.Fields("id", "$['payload']['body']", "threadId"),
				NextPage:   "08277485409175924556",
				Since:      time.Date(2024, 9, 31, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					// Get collection of all messages. This is a list of identifiers.
					If: mockcond.And{
						mockcond.Path("/gmail/v1/users/me/messages"),
						mockcond.QueryParam("maxResults", "500"), // default page size
						mockcond.QueryParam("pageToken", "08277485409175924556"),
						mockcond.QueryParam("q", "after:2024/10/01 before:2026/01/08"),
					},
					Then: mockserver.Response(http.StatusOK, responseMessagesLastPage),
				}, {
					// Each message is fetched.
					// The last page is a collection of messages with just 1 message where id=19174f3eeda702ed.
					If:   mockcond.Path("/gmail/v1/users/me/messages/19174f3eeda702ed"),
					Then: mockserver.Response(http.StatusOK, responseMessageItem),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"payload": map[string]any{
							"body": map[string]any{
								"size": float64(0), // nested field was requested
							},
						},
						"id":       "19174f3eeda702ed",
						"threadid": "19174f3eeda702ed", // fields are lower case
					},
					Raw: map[string]any{
						"id":       "19174f3eeda702ed",
						"threadId": "19174f3eeda702ed",
						// Message content is embedded.
						"snippet":      "Restart your 14-day free trial now.",
						"sizeEstimate": float64(29817),
						"historyId":    "31772",
						"internalDate": "1769523000000",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read drafts with embedded messages",
			Input: common.ReadParams{
				ObjectName: "drafts",
				Fields: connectors.Fields(
					"$['message']['snippet']",
					"$['message']['payload']['body']",
					"$['message']['payload']['mimeType']",
				),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/gmail/v1/users/me/drafts"),
						mockcond.QueryParam("maxResults", "500"),
					},
					Then: mockserver.Response(http.StatusOK, responseDrafts),
				}, {
					If:   mockcond.Path("/gmail/v1/users/me/messages/19c5a57884cc0fc0"),
					Then: mockserver.Response(http.StatusOK, responseDraftMessageItem1),
				}, {
					If:   mockcond.Path("/gmail/v1/users/me/messages/19c5a576fbc9a5ec"),
					Then: mockserver.Response(http.StatusOK, responseDraftMessageItem2),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "r5482888295841858934",
					Fields: map[string]any{
						"message": map[string]any{
							"snippet": "(DRAFT) Good news. You are being upgraded from Basic",
							"payload": map[string]any{
								"mimetype": "multipart/alternative",
								"body": map[string]any{
									"size": float64(0),
								},
							},
						},
					},
					Raw: map[string]any{
						"id": "r5482888295841858934",
						"message": map[string]any{
							"id":       "19c5a57884cc0fc0",
							"threadId": "19c2643cde3cab88",
							"labelIds": []any{"DRAFT"},
							"snippet":  "(DRAFT) Good news. You are being upgraded from Basic",
							"payload": map[string]any{
								"partId":   "",
								"mimeType": "multipart/alternative",
								"filename": "",
								"headers": []any{
									map[string]any{
										"name":  "Subject",
										"value": "(DRAFT) Good news",
									},
								},
								"body": map[string]any{
									"size": float64(0),
								},
								"parts": []any{},
							},
							"sizeEstimate": float64(1924),
							"historyId":    "39895",
							"internalDate": "1771042211000",
						},
					},
				}, {
					Id: "r9115380863326367925",
					Fields: map[string]any{
						"message": map[string]any{
							"snippet": "(DRAFT) You have had Linux installed for about a month now",
							"payload": map[string]any{
								"mimetype": "multipart/alternative",
								"body": map[string]any{
									"size": float64(0),
								},
							},
						},
					},
					Raw: map[string]any{
						"id": "r9115380863326367925",
						"message": map[string]any{
							"id":       "19c5a576fbc9a5ec",
							"threadId": "19c5a5623749ebbb",
							"labelIds": []any{"DRAFT"},
							"snippet":  "(DRAFT) You have had Linux installed for about a month now",
							"payload": map[string]any{
								"partId":   "",
								"mimeType": "multipart/alternative",
								"filename": "",
								"headers": []any{
									map[string]any{
										"name":  "Subject",
										"value": "(DRAFT) Linux installed",
									},
								},
								"body": map[string]any{
									"size": float64(0),
								},
								"parts": []any{},
							},
							"sizeEstimate": float64(1677),
							"historyId":    "39887",
							"internalDate": "1771042205000",
						},
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
				return constructTestMailConnector(tt.Server.URL)
			})
		})
	}
}
