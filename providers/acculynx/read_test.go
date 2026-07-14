package acculynx

import (
	_ "embed"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

//go:embed test/read/users-first-page.json
var usersFirstPageResponse []byte

//go:embed test/read/users-last-page.json
var usersLastPageResponse []byte

//go:embed test/read/jobs-list.json
var jobsListResponse []byte

//go:embed test/read/job-contacts-001.json
var jobContacts001Response []byte

//go:embed test/read/job-contacts-002.json
var jobContacts002Response []byte

//go:embed test/read/calendar-appointments.json
var calendarAppointmentsResponse []byte

//go:embed test/read/calendars-list.json
var calendarsListResponse []byte

//go:embed test/read/units-of-measure.json
var unitsOfMeasureResponse []byte

//go:embed test/read/custom-fields/definitions.json
var customFieldDefinitionsFixture []byte

//go:embed test/read/custom-fields/definitions-empty.json
var customFieldDefinitionsEmptyResponse []byte

//go:embed test/read/custom-fields/contacts-list.json
var customFieldsContactsListResponse []byte

//go:embed test/read/custom-fields/contact-001-values.json
var customFieldsContact001ValuesResponse []byte

//go:embed test/read/custom-fields/contact-002-values.json
var customFieldsContact002ValuesResponse []byte

//go:embed test/read/jobs-list-single.json
var jobsListSingleResponse []byte

//go:embed test/read/job-invoices-page1.json
var jobInvoicesPage1Response []byte

//go:embed test/read/job-invoices-page2.json
var jobInvoicesPage2Response []byte

//go:embed test/read/job-history.json
var jobHistoryResponse []byte

//go:embed test/read/jobs-with-contacts.json
var jobsWithContactsResponse []byte

//go:embed test/read/contacts-includes.json
var contactsIncludesResponse []byte

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read contacts always requests emailAddress,phoneNumber includes",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				PageSize:   100,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/contacts"),
							mockcond.QueryParam("includes", "emailAddress,phoneNumber"),
						},
						Then: mockserver.Response(http.StatusOK, contactsIncludesResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "ctc-100"},
					Raw:    map[string]any{"id": "ctc-100"},
				}},
				Done: true,
			},
		},
		{
			Name: "Read jobs with contacts association attaches embedded contacts",
			Input: common.ReadParams{
				ObjectName:        "jobs",
				Fields:            connectors.Fields("id"),
				AssociatedObjects: []string{"contacts"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("includes", "contacts"),
						},
						Then: mockserver.Response(http.StatusOK, jobsWithContactsResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "job-001"},
						Raw:    map[string]any{"id": "job-001"},
						Associations: map[string][]common.Association{
							"contacts": {{
								ObjectId:                    "ctc-100",
								Raw:                         map[string]any{"id": "ctc-100", "firstName": "Diane"},
								ProviderAssociationMetadata: map[string]any{"isPrimary": true},
							}},
						},
					},
					{
						Fields: map[string]any{"id": "job-002"},
						Raw:    map[string]any{"id": "job-002"},
						Associations: map[string][]common.Association{
							"contacts": {{
								ObjectId:                    "ctc-200",
								Raw:                         map[string]any{"id": "ctc-200"},
								ProviderAssociationMetadata: map[string]any{"isPrimary": true},
							}},
						},
					},
				},
				Done: true,
			},
		},
		{
			Name: "Read jobs without association omits includes and attaches no associations",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParamsMissing("includes"),
						},
						Then: mockserver.Response(http.StatusOK, jobsWithContactsResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"id": "job-001"}, Raw: map[string]any{"id": "job-001"}},
					{Fields: map[string]any{"id": "job-002"}, Raw: map[string]any{"id": "job-002"}},
				},
				Done: true,
			},
		},
		{
			Name: "Object must be supported",
			Input: common.ReadParams{
				ObjectName: "nonexistent",
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Read users full page returns next page",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "displayName"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/users"),
							mockcond.QueryParam("pageSize", "2"),
							mockcond.QueryParam("recordStartIndex", "0"),
						},
						Then: mockserver.Response(http.StatusOK, usersFirstPageResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					// Fields keys are lowercased by the framework; Raw preserves
					// the original camelCase from the API response.
					Fields: map[string]any{
						"id":          "u1",
						"displayname": "Alice Anderson",
					},
					// Raw must include fields the caller didn't request
					// (firstName/lastName/email/status), proving the original
					// payload is preserved through the read pipeline.
					Raw: map[string]any{
						"id":          "u1",
						"displayName": "Alice Anderson",
						"firstName":   "Alice",
						"lastName":    "Anderson",
						"email":       "alice@example.com",
						"status":      "Active",
					},
				}},
				NextPage: testconn.URLTestServer + "/api/v2/users?pageSize=2&recordStartIndex=2",
				Done:     false,
			},
		},
		{
			Name: "Read users partial page signals Done",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
				NextPage:   testconn.URLTestServer + "/api/v2/users?pageSize=2&recordStartIndex=4",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.QueryParam("recordStartIndex", "4"),
						},
						Then: mockserver.Response(http.StatusOK, usersLastPageResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read jobs with Since adds provider-side ModifiedDate filter",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("dateFilterType", "ModifiedDate"),
							mockcond.QueryParam("startDate", "2026-04-01"),
							mockcond.QueryParam("endDate", "2026-04-30"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				Done: true,
			},
		},
		{
			Name: "Read contacts flattens custom field values onto each row",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields: connectors.Fields(
					"id", "firstName", "customer_preference", "preferred_contact_method",
				),
				PageSize: 100,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/contacts"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldsContactsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/company-settings/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsFixture),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/contacts/ctc_001/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldsContact001ValuesResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/contacts/ctc_002/custom-fields"),
						},
						Then: mockserver.Response(http.StatusOK, customFieldsContact002ValuesResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":        "ctc_001",
							"firstname": "Carol",
							// customer_preference is a custom field — single-value
							// arrays are unwrapped to scalars. Slug derives from
							// the definition label lowercased, spaces→underscores.
							"customer_preference": "White Roof",
							// preferred_contact_method is a custom field with
							// multiple values — preserved as a slice.
							"preferred_contact_method": []string{"Phone", "Email"},
						},
						// Raw must remain the untouched API response: custom
						// field values must NOT bleed into Raw, and
						// provider-returned fields like _link must NOT be
						// stripped (paranoia check).
						Raw: map[string]any{
							"id":           "ctc_001",
							"_link":        "https://api.acculynx.com/api/v2/contacts/ctc_001",
							"firstName":    "Carol",
							"lastName":     "Customer",
							"modifiedDate": "2026-04-10T12:00:00Z",
						},
					},
					{
						// ctc_002 has no custom-field values — built-in fields
						// only.
						Fields: map[string]any{
							"id":        "ctc_002",
							"firstname": "Dave",
						},
						Raw: map[string]any{
							"id":           "ctc_002",
							"_link":        "https://api.acculynx.com/api/v2/contacts/ctc_002",
							"firstName":    "Dave",
							"lastName":     "Donor",
							"modifiedDate": "2026-04-11T12:00:00Z",
						},
					},
				},
				Done: true,
			},
		},
		{
			Name: "Read jobs/contacts fans out per job and flattens results",
			Input: common.ReadParams{
				ObjectName: "jobs/contacts",
				Fields:     connectors.Fields("id", "firstName", "lastName", "jobId"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("recordStartIndex", "0"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/contacts"),
						},
						Then: mockserver.Response(http.StatusOK, jobContacts001Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-002/contacts"),
						},
						Then: mockserver.Response(http.StatusOK, jobContacts002Response),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 3,
				Done: true,
			},
		},
		{
			Name: "Read jobs/invoices follows NextPage per parent (paginated fan-out)",
			Input: common.ReadParams{
				ObjectName: "jobs/invoices",
				Fields:     connectors.Fields("id", "balanceDue"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListSingleResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/invoices"),
							mockcond.QueryParam("pageStartIndex", "0"),
						},
						Then: mockserver.Response(http.StatusOK, jobInvoicesPage1Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/invoices"),
							mockcond.QueryParam("pageStartIndex", "2"),
						},
						Then: mockserver.Response(http.StatusOK, jobInvoicesPage2Response),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 3,
				Done: true,
			},
		},
		{
			Name: "Read jobs/history pushes Since/Until to server-side startDate/endDate (date-only)",
			Input: common.ReadParams{
				ObjectName: "jobs/history",
				Fields:     connectors.Fields("action", "date"),
				Since:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListSingleResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/history"),
							mockcond.QueryParam("startDate", "2026-05-01"),
							mockcond.QueryParam("endDate", "2026-05-15"),
						},
						Then: mockserver.Response(http.StatusOK, jobHistoryResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				Done: true,
			},
		},
		{
			Name: "Read calendars/appointments defaults a 30-day window",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id", "title"),
				PageSize:   100,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/calendars"),
						},
						Then: mockserver.Response(http.StatusOK, calendarsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/calendars/cal-001/appointments"),
						},
						Then: mockserver.Response(http.StatusOK, calendarAppointmentsResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
			},
		},
		{
			Name: "Read acculynx/units-of-measure honours custom responseKey",
			Input: common.ReadParams{
				ObjectName: "acculynx/units-of-measure",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/acculynx/units-of-measure"),
						},
						Then: mockserver.Response(http.StatusOK, unitsOfMeasureResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestReadConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestReadConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: &http.Client{},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
