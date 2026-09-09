package reply

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseContactsFirstPage := testutils.DataFromFile(t, "read/contacts-first-page.json")
	responseContactsLastPage := testutils.DataFromFile(t, "read/contacts.json")
	responseCustomFields := testutils.DataFromFile(t, "read/custom-fields.json")
	responseSchedules := testutils.DataFromFile(t, "read/schedules.json")
	responseSequences := testutils.DataFromFile(t, "read/sequences.json")
	errorValidation := testutils.DataFromFile(t, "read/err-validation.json")
	errorUnauthorized := testutils.DataFromFile(t, "read/err-unauthorized.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contacts", Fields: nil},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is rejected by the registry",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Contacts first page paginates by skip and resolves custom field titles",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v3/contacts"),
						mockcond.QueryParam("top", "100"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsFirstPage),
				}, {
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "761575310",
					Fields: map[string]any{
						"email": "john.roe@example.com",
					},
					Raw: map[string]any{
						"firstName": "John",
						"lastName":  "Roe",
					},
				}},
				NextPage: testconn.URLTestServer + "/v3/contacts?skip=100&top=100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Contacts last page resolves custom field titles and finishes",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email", "Deal Stage"),
				NextPage:   common.NextPageToken(testconn.URLTestServer + "/v3/contacts?skip=100&top=100"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.Path("/v3/contacts"),
						mockcond.QueryParam("skip", "100"),
					},
					Then: mockserver.Response(http.StatusOK, responseContactsLastPage),
				}, {
					If:   mockcond.Path("/v3/custom-fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFields),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "759458702",
					Fields: map[string]any{
						"email":      "jane.doe@example.com",
						"deal stage": "probe",
					},
					// Raw keeps the provider's key/value shape untouched.
					Raw: map[string]any{
						"customFields": []any{map[string]any{
							"key":   "150697",
							"value": "probe",
						}},
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Schedules is a single root-level array page",
			Input: common.ReadParams{
				ObjectName: "schedules",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/schedules"),
				Then:  mockserver.Response(http.StatusOK, responseSchedules),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "559907",
					Fields: map[string]any{
						"name": "Default",
					},
					Raw: map[string]any{
						"timezoneId": "SA Pacific Standard Time",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Sequences incremental read sends createdAfter",
			Input: common.ReadParams{
				ObjectName: "sequences",
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v3/sequences"),
					mockcond.QueryParam("createdAfter", "2026-01-02T03:04:05Z"),
				},
				Then: mockserver.ResponseString(http.StatusOK, `{"items":[],"hasMore":false}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 0,
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Sequences Since and Until window filters connector-side",
			Input: common.ReadParams{
				ObjectName: "sequences",
				Fields:     connectors.Fields("name"),
				Since:      time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC),
				// Between the two records' created times: only the earlier
				// sequence (12:15:35) falls inside the window.
				Until: time.Date(2026, 9, 8, 13, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v3/sequences"),
				Then:  mockserver.Response(http.StatusOK, responseSequences),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id: "1768686",
					Fields: map[string]any{
						"name": "Deep Connector Test Sequence",
					},
					Raw: map[string]any{
						"created": "2026-09-08T12:15:35.52+00:00",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Unauthorized problem response",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("email"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, errorUnauthorized),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("Unauthorized: Authentication credentials are missing or invalid."),
			},
		},
		{
			Name: "Validation problem response carries pointer detail",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorValidation),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Validation failed: /top Value has an invalid type."),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
