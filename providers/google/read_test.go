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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorPageToken := testutils.DataFromFile(t, "read/error-page-token.json")
	responseCalendarListFirstPage := testutils.DataFromFile(t, "read/calendarList/1-first-page.json")
	responseCalendarListLastPage := testutils.DataFromFile(t, "read/calendarList/2-last-page.json")
	responseSettingsFirstPage := testutils.DataFromFile(t, "read/settings/1-first-page.json")
	responseSettingsLastPage := testutils.DataFromFile(t, "read/settings/2-last-page.json")

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
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "orders", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
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
			Name: "Read calendarList first page",
			Input: common.ReadParams{
				ObjectName: "calendarList",
				Fields:     connectors.Fields("summary", "description"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseCalendarListFirstPage),
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
				NextPage: testroutines.URLTestServer + "/calendar/v3/users/me/calendarList?maxResults=100" +
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
					mockcond.PathSuffix("calendar/v3/users/me/calendarList"),
					mockcond.QueryParam("maxResults", "100"),
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
				If:    mockcond.PathSuffix("calendar/v3/users/me/settings"),
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
				NextPage: testroutines.URLTestServer + "/calendar/v3/users/me/settings?maxResults=100" +
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
				If:    mockcond.PathSuffix("calendar/v3/users/me/settings"),
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
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithModule(ModuleCalendar),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
