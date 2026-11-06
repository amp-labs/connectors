package acculynx

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"gotest.tools/v3/assert"
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

//go:embed test/read/job-financials-001.json
var jobFinancials001Response []byte

//go:embed test/read/job-financials-002.json
var jobFinancials002Response []byte

//go:embed test/read/calendar-appointments.json
var calendarAppointmentsResponse []byte

//go:embed test/read/calendar-appointments-assoc.json
var calendarAppointmentsAssocResponse []byte

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

//go:embed test/read/jobs-with-initial-appointment.json
var jobsWithInitialAppointmentResponse []byte

//go:embed test/read/job-representatives.json
var jobRepresentativesResponse []byte

//go:embed test/read/estimates-list.json
var estimatesListResponse []byte

//go:embed test/read/estimate-detail-est-1.json
var estimateDetailEst1Response []byte

//go:embed test/read/estimate-detail-est-2.json
var estimateDetailEst2Response []byte

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
							mockcond.QueryParam("includes", "initialAppointment,contacts"),
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
								ProviderAssociationMetadata: map[string]any{"isPrimary": true},
							}},
						},
					},
				},
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment%2Ccontacts&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			Name: "Read jobs without association still requests initialAppointment and attaches no associations",
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
							mockcond.QueryParam("includes", "initialAppointment"),
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
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// The initialAppointment expansion is what makes startDate/endDate/notes
			// present at all — without ?includes=initialAppointment AccuLynx omits the
			// property from the job payload entirely.
			Name: "Read jobs surfaces the expanded initialAppointment object",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id", "initialAppointment"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("includes", "initialAppointment"),
						},
						Then: mockserver.Response(http.StatusOK, jobsWithInitialAppointmentResponse),
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
					Fields: map[string]any{
						"id": "875b28e8-4f10-44e2-9af1-aff90527291e",
						"initialappointment": map[string]any{
							"startDate": "2025-03-20T16:30:00Z",
							"endDate":   "2025-03-20T17:00:00Z",
							"notes":     "Customer contact information:\nFoobar - McHealy, David\n",
							"_link": "https://api.acculynx.com/api/v2/jobs/" +
								"875b28e8-4f10-44e2-9af1-aff90527291e/initial-appointment",
						},
					},
					Raw: map[string]any{
						"id": "875b28e8-4f10-44e2-9af1-aff90527291e",
						"initialAppointment": map[string]any{
							"startDate": "2025-03-20T16:30:00Z",
							"endDate":   "2025-03-20T17:00:00Z",
							"notes":     "Customer contact information:\nFoobar - McHealy, David\n",
							"_link": "https://api.acculynx.com/api/v2/jobs/" +
								"875b28e8-4f10-44e2-9af1-aff90527291e/initial-appointment",
						},
					},
				}},
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
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
							mockcond.QueryParam("pageStartIndex", "0"),
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
				NextPage: testconn.URLTestServer + "/api/v2/users?pageSize=2&pageStartIndex=2",
				Done:     false,
			},
		},
		{
			Name: "Read users partial page signals Done",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
				NextPage:   testconn.URLTestServer + "/api/v2/users?pageSize=2&pageStartIndex=4",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.QueryParam("pageStartIndex", "4"),
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
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&dateFilterType=ModifiedDate&endDate=2026-04-30&includes=initialAppointment&pageSize=25&pageStartIndex=0&startDate=2026-04-01",
				Done: false,
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
			// The financials endpoint answers with one object per job, not a
			// list; each response becomes exactly one row. The leaf request
			// must ask for the worksheet and amendments expansions.
			Name: "Read jobs/financials wraps each job's object as one row",
			Input: common.ReadParams{
				ObjectName: "jobs/financials",
				Fields:     connectors.Fields("jobId", "balanceDue"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("pageStartIndex", "0"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/financials"),
							mockcond.QueryParam("includes", "worksheet,amendments"),
						},
						Then: mockserver.Response(http.StatusOK, jobFinancials001Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-002/financials"),
							mockcond.QueryParam("includes", "worksheet,amendments"),
						},
						Then: mockserver.Response(http.StatusOK, jobFinancials002Response),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"jobid": "job-001", "balancedue": -2400.00},
						Raw:    map[string]any{"jobId": "job-001", "balanceDue": -2400.00},
					},
					{
						Fields: map[string]any{"jobid": "job-002", "balancedue": 150.75},
						Raw:    map[string]any{"jobId": "job-002", "balanceDue": 150.75},
					},
				},
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			Name: "Read jobs/financials attaches the job association",
			Input: common.ReadParams{
				ObjectName:        "jobs/financials",
				Fields:            connectors.Fields("jobId", "balanceDue"),
				AssociatedObjects: []string{"jobs"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("pageStartIndex", "0"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/financials"),
						},
						Then: mockserver.Response(http.StatusOK, jobFinancials001Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-002/financials"),
						},
						Then: mockserver.Response(http.StatusOK, jobFinancials002Response),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"jobid": "job-001", "balancedue": -2400.00},
						Raw:    map[string]any{"jobId": "job-001", "balanceDue": -2400.00},
						Associations: map[string][]common.Association{
							"jobs": {{ObjectId: "job-001"}},
						},
					},
					{
						Fields: map[string]any{"jobid": "job-002", "balancedue": 150.75},
						Raw:    map[string]any{"jobId": "job-002", "balanceDue": 150.75},
						Associations: map[string][]common.Association{
							"jobs": {{ObjectId: "job-002"}},
						},
					},
				},
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// Second half of the two-population jobs sweep: a short page of the
			// unassigned listing ends the read with no further hand-over.
			Name: "Jobs unassigned sweep ends the read",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("assignment", "unassigned"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListResponse),
					},
					{
						If:   mockcond.Path("/api/v2/company-settings/custom-fields"),
						Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "",
				Done:     true,
			},
		},
		{
			// AccuLynx rejects a page that ends past position 100000, so a list
			// with more matches must be read sorted by modifiedDate.
			Name: "Read jobs with more than 100000 matches restarts the list sorted by modifiedDate",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
			},
			Server: jobsServer(map[string]string{"startDate": "2026-01-01", "pageStartIndex": "0"},
				jobsPage(150000, 0, 25, "2026-03-01T10:00:00Z", "2026-02-01T10:00:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06&includes=initialAppointment" +
					"&pageSize=25&pageStartIndex=0&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-01-01"),
			},
		},
		{
			// The default order (by created date) does not move when a job is
			// edited during the read, so it is kept whenever it can reach the end.
			Name: "Read jobs with 100000 matches or fewer keeps the default order",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
			},
			Server: jobsServer(map[string]string{"startDate": "2026-01-01", "pageStartIndex": "0"},
				jobsPage(247, 0, 25, "2026-03-01T10:00:00Z", "2026-02-01T10:00:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06&includes=initialAppointment" +
					"&pageSize=25&pageStartIndex=25&startDate=2026-01-01"),
			},
		},
		{
			Name: "Read jobs at the 100000 limit moves startDate to the day before the newest job",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=99975" +
					"&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-01-01"),
			},
			Server: jobsServer(map[string]string{"pageStartIndex": "99975"},
				jobsPage(150000, 99975, 25, "2026-07-14T08:00:00Z", "2026-07-14T09:30:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06&includes=initialAppointment" +
					"&pageSize=25&pageStartIndex=0&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-07-13"),
			},
		},
		{
			// 99999 is under 100000, but 99999 + 3 is not: the limit is on where
			// the page ends, not where it starts.
			Name: "Read jobs with pageSize 3 moves startDate before a page that would end past 100000",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
				PageSize:   3,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=3&pageStartIndex=99996" +
					"&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-01-01"),
			},
			Server: jobsServer(map[string]string{"pageStartIndex": "99996"},
				jobsPage(150000, 99996, 3, "2026-07-14T08:00:00Z", "2026-07-14T09:30:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 3,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06&includes=initialAppointment" +
					"&pageSize=3&pageStartIndex=0&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-07-13"),
			},
		},
		{
			// A cursor saved before this fix points at 100000 and only ever gets
			// a 400. It must not be sent as is: the server below answers only
			// the sorted restart.
			Name: "Read jobs from a saved cursor past the limit restarts its list sorted",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=100000&startDate=2026-01-01"),
			},
			Server: jobsServer(
				map[string]string{"sortBy": "ModifiedDate", "sortOrder": "Ascending", "pageStartIndex": "0"},
				jobsPage(150000, 0, 25, "2026-01-02T10:00:00Z", "2026-01-03T10:00:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06&includes=initialAppointment" +
					"&pageSize=25&pageStartIndex=25&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-01-01"),
			},
		},
		{
			// With no Since/Until the URL has no endDate yet, so today is used.
			Name: "Read jobs without Since/Until adds a date filter at the 100000 limit",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				NextPage: jobsURL("includes=initialAppointment&pageSize=25" +
					"&pageStartIndex=99975&sortBy=ModifiedDate&sortOrder=Ascending"),
			},
			Server: jobsServer(map[string]string{"pageStartIndex": "99975"},
				jobsPage(150000, 99975, 25, "2026-07-14T08:00:00Z", "2026-07-14T09:30:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=" + time.Now().UTC().Format(time.DateOnly) +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=0" +
					"&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-07-13"),
			},
		},
		{
			// The new startDate (07-13) is not after the current one (07-13):
			// more than 100000 jobs were modified within two days.
			Name: "Read jobs past the limit with too many jobs in two days returns an error",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=99975" +
					"&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-07-13"),
			},
			Server: jobsServer(map[string]string{"pageStartIndex": "99975"},
				jobsPage(150000, 99975, 25, "2026-07-14T08:00:00Z", "2026-07-14T09:30:00Z")),
			ExpectedErrs: []error{errJobsWindowStalled},
		},
		{
			Name: "Read jobs at the limit with an unsorted page returns an error",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=99975" +
					"&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-01-01"),
			},
			Server: jobsServer(map[string]string{"pageStartIndex": "99975"},
				jobsPage(150000, 99975, 25, "2026-07-14T09:30:00Z", "2026-03-02T10:00:00Z")),
			ExpectedErrs: []error{errJobsNotOrdered},
		},
		{
			// The unassigned jobs are a new list: they start from the read's own
			// dates in the default order, not from where the assigned list ended.
			Name: "Jobs unassigned sweep restores the read's dates and default order",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC),
				NextPage: jobsURL("dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=50" +
					"&sortBy=ModifiedDate&sortOrder=Ascending&startDate=2026-07-13"),
			},
			Server: jobsServer(map[string]string{"pageStartIndex": "50"},
				jobsPage(75, 50, 25, "2026-08-01T10:00:00Z", "2026-08-02T10:00:00Z")),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: jobsURL("assignment=unassigned&dateFilterType=ModifiedDate&endDate=2026-11-06" +
					"&includes=initialAppointment&pageSize=25&pageStartIndex=0&startDate=2026-01-01"),
			},
		},
		{
			// An unassigned job has no representatives and its leaf answers
			// 404; the fan-out must treat that parent as having no child
			// records rather than failing the read.
			Name: "Fan-out skips a parent whose leaf answers 404",
			Input: common.ReadParams{
				ObjectName: "jobs/representatives",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs"),
							mockcond.QueryParam("pageStartIndex", "0"),
						},
						Then: mockserver.Response(http.StatusOK, jobsListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-001/representatives"),
						},
						Then: mockserver.Response(http.StatusOK, jobRepresentativesResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/jobs/job-002/representatives"),
						},
						Then: mockserver.ResponseString(http.StatusNotFound, `{"title":"Not Found","status":404}`),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
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
							mockcond.QueryParam("pageStartIndex", "0"),
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
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
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
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=2&pageStartIndex=0",
				Done: false,
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
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// No Since/Until: the default window runs from 30 days back to 60
			// days ahead (see TestAppointmentWindows for the exact dates).
			Name: "Read calendars/appointments reads the default window",
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
			Name: "Read calendars/appointments attaches job and user associations",
			Input: common.ReadParams{
				ObjectName:        "calendars/appointments",
				Fields:            connectors.Fields("id", "title"),
				PageSize:          100,
				AssociatedObjects: []string{"jobs", "users"},
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
						Then: mockserver.Response(http.StatusOK, calendarAppointmentsAssocResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						// Job-linked event: both edges. calendarId (cal-001) is the
						// fan-out parent -> user; jobId rides on the body -> job.
						Fields: map[string]any{"id": "appt-1", "title": "Site visit"},
						Raw:    map[string]any{"id": "appt-1", "jobId": "job-777"},
						Associations: map[string][]common.Association{
							"jobs":  {{ObjectId: "job-777"}},
							"users": {{ObjectId: "cal-001"}},
						},
					},
					{
						// Personal event: no jobId, so only the user edge is emitted.
						Fields: map[string]any{"id": "appt-2", "title": "Lunch"},
						Raw:    map[string]any{"id": "appt-2"},
						Associations: map[string][]common.Association{
							"users": {{ObjectId: "cal-001"}},
						},
					},
				},
				Done: true,
			},
		},
		{
			Name: "Read calendars/appointments without association request attaches nothing",
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
						Then: mockserver.Response(http.StatusOK, calendarAppointmentsAssocResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"id": "appt-1"}, Raw: map[string]any{"id": "appt-1"}},
					{Fields: map[string]any{"id": "appt-2"}, Raw: map[string]any{"id": "appt-2"}},
				},
				Done: true,
			},
		},
		{
			Name: "Read calendars/appointments sends Since/Until as one date window",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 4, 1, 8, 30, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 23, 59, 0, 0, time.UTC),
			},
			Server: appointmentsServer(appointmentsWindowCase("2026-04-01", "2026-04-30",
				appointmentsPage(`{"id": "appt-1", "start": "2026-04-15T14:00:00Z"}`))),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 1, Done: true},
		},
		{
			// AccuLynx rejects a range of 90 days or more, so a 120-day read is
			// two requests: 89 days, then the remainder.
			Name: "Read calendars/appointments splits a range of 90 days or more into windows",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: appointmentsServer(
				appointmentsWindowCase("2026-01-01", "2026-03-31",
					appointmentsPage(`{"id": "appt-1", "start": "2026-02-10T14:00:00Z"}`)),
				appointmentsWindowCase("2026-04-01", "2026-04-30",
					appointmentsPage(`{"id": "appt-2", "start": "2026-04-20T14:00:00Z"}`)),
			),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 2, Done: true},
		},
		{
			// A multi-day CrewLabor event comes back as one row per day, all
			// sharing the event id: every day is kept. The row for 03-31 is
			// returned by both windows (the API matches the day boundary
			// loosely) and is emitted once.
			Name: "Read calendars/appointments keeps each day of a multi-day event and drops window repeats",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id", "start"),
				Since:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: appointmentsServer(
				appointmentsWindowCase("2026-01-01", "2026-03-31",
					appointmentsPage(`{"id": "crew-1", "start": "2026-03-31T04:00:00Z", "allDay": true}`)),
				appointmentsWindowCase("2026-04-01", "2026-04-30",
					appointmentsPage(
						`{"id": "crew-1", "start": "2026-03-31T04:00:00Z", "allDay": true}`,
						`{"id": "crew-1", "start": "2026-04-01T04:00:00Z", "allDay": true}`,
						`{"id": "crew-1", "start": "2026-04-02T04:00:00Z", "allDay": true}`,
					)),
			),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z"},
						Raw:    map[string]any{"id": "crew-1", "start": "2026-03-31T04:00:00Z", "allDay": true},
					},
					{
						Fields: map[string]any{"id": "crew-1", "start": "2026-04-01T04:00:00Z"},
						Raw:    map[string]any{"id": "crew-1", "start": "2026-04-01T04:00:00Z", "allDay": true},
					},
					{
						Fields: map[string]any{"id": "crew-1", "start": "2026-04-02T04:00:00Z"},
						Raw:    map[string]any{"id": "crew-1", "start": "2026-04-02T04:00:00Z", "allDay": true},
					},
				},
				Done: true,
			},
		},
		{
			Name: "Read calendars/appointments returns the API's date range error",
			Input: common.ReadParams{
				ObjectName: "calendars/appointments",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Until:      time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			},
			Server: appointmentsServer(appointmentsWindowCase("2026-04-01", "2026-04-30",
				mockserver.ResponseString(http.StatusBadRequest, `{
					"type": "https://tools.ietf.org/html/rfc9110#section-15.5.1",
					"title": "Range between Start Date and End Date must be less than 90 days.",
					"status": 400,
					"detail": "BadRequest"
				}`))),
			ExpectedErrs: []error{common.ErrCaller},
		},
		{
			Name: "Read jobs/representatives attaches the parent job association",
			Input: common.ReadParams{
				ObjectName:        "jobs/representatives",
				Fields:            connectors.Fields("id", "type"),
				AssociatedObjects: []string{"jobs"},
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
							mockcond.Path("/api/v2/jobs/job-001/representatives"),
						},
						Then: mockserver.Response(http.StatusOK, jobRepresentativesResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						// The representative body carries no jobId — the edge comes
						// from the fan-out parent (job-001), like appointment->user.
						Fields: map[string]any{"id": "rep-1", "type": "SalesOwner"},
						Raw:    map[string]any{"id": "rep-1", "type": "SalesOwner"},
						Associations: map[string][]common.Association{
							"jobs": {{ObjectId: "job-001"}},
						},
					},
					{
						Fields: map[string]any{"id": "rep-2", "type": "ArOwner"},
						Raw:    map[string]any{"id": "rep-2", "type": "ArOwner"},
						Associations: map[string][]common.Association{
							"jobs": {{ObjectId: "job-001"}},
						},
					},
				},
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			Name: "Read jobs/representatives without association request attaches nothing",
			Input: common.ReadParams{
				ObjectName: "jobs/representatives",
				Fields:     connectors.Fields("id", "type"),
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
							mockcond.Path("/api/v2/jobs/job-001/representatives"),
						},
						Then: mockserver.Response(http.StatusOK, jobRepresentativesResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{Fields: map[string]any{"id": "rep-1"}, Raw: map[string]any{"id": "rep-1"}},
					{Fields: map[string]any{"id": "rep-2"}, Raw: map[string]any{"id": "rep-2"}},
				},
				// The assigned sweep is exhausted; the read continues into the
				// unassigned population before it is Done.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// The /estimates list returns stubs only (id, isPrimary, job) and
			// ignores ?includes=; the substantive fields must come from one
			// GET /estimates/{id} per record. Hydration requires the
			// server-set opt-in.
			Name: "Read estimates with opts hydrates each row from the detail endpoint",
			Input: common.ReadParams{
				ObjectName: "estimates",
				Fields:     connectors.Fields("id", "estimateNumber", "createdDate", "financials"),
				Opts:       ReadParamsOpts{HydrateEstimates: true},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates"),
						},
						Then: mockserver.Response(http.StatusOK, estimatesListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates/est-1"),
							mockcond.QueryParam("includes", "createdBy,modifiedBy,sections"),
						},
						Then: mockserver.Response(http.StatusOK, estimateDetailEst1Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates/est-2"),
							mockcond.QueryParam("includes", "createdBy,modifiedBy,sections"),
						},
						Then: mockserver.Response(http.StatusOK, estimateDetailEst2Response),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						// The detail is provider data, so it lands in Raw as
						// well as Fields — raw stays "the object as AccuLynx
						// gave it", assembled from the list and detail calls.
						Fields: map[string]any{
							"id":             "est-1",
							"estimatenumber": "1042",
							"createddate":    "2026-05-10T17:56:54Z",
							"financials": map[string]any{
								"taxRate":       0.0,
								"taxTotal":      0.0,
								"overheadRate":  0.0,
								"overheadTotal": 0.0,
								"profitRate":    0.0,
								"profitTotal":   0.0,
								"totalCost":     100.0,
								"totalPrice":    142.86,
							},
						},
						Raw: map[string]any{
							"id":             "est-1",
							"isPrimary":      true,
							"estimateNumber": "1042",
							"createdDate":    "2026-05-10T17:56:54Z",
							"title":          "Roof replacement",
						},
					},
					{
						Fields: map[string]any{
							"id":             "est-2",
							"estimatenumber": "1043",
							"createddate":    "2026-06-01T08:15:00Z",
						},
						Raw: map[string]any{
							"id":             "est-2",
							"isPrimary":      false,
							"estimateNumber": "1043",
							"title":          "Gutter repair",
						},
					},
				},
				Done: true,
			},
		},
		{
			// No Opts set: the association is a pure reshape of the list
			// payload. The mock serves only the list — any hydration call
			// would hit the 500 default and fail the read.
			Name: "Read estimates with jobs association attaches the job edge without hydration calls",
			Input: common.ReadParams{
				ObjectName:        "estimates",
				Fields:            connectors.Fields("id"),
				AssociatedObjects: []string{"jobs"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates"),
						},
						Then: mockserver.Response(http.StatusOK, estimatesListResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "est-1"},
						Raw:    map[string]any{"id": "est-1"},
						Associations: map[string][]common.Association{
							"jobs": {{
								ObjectId:                    "job-777",
								ProviderAssociationMetadata: map[string]any{"isPrimary": true},
							}},
						},
					},
					{
						Fields: map[string]any{"id": "est-2"},
						Raw:    map[string]any{"id": "est-2"},
						Associations: map[string][]common.Association{
							"jobs": {{
								ObjectId:                    "job-888",
								ProviderAssociationMetadata: map[string]any{"isPrimary": false},
							}},
						},
					},
				},
				Done: true,
			},
		},
		{
			// An estimate deleted between the list and detail calls must not sink
			// the read — the row keeps its thin list shape.
			Name: "Read estimates keeps the thin row when its detail returns 404",
			Input: common.ReadParams{
				ObjectName: "estimates",
				Fields:     connectors.Fields("id", "estimateNumber"),
				Opts:       ReadParamsOpts{HydrateEstimates: true},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates"),
						},
						Then: mockserver.Response(http.StatusOK, estimatesListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates/est-1"),
							mockcond.QueryParam("includes", "createdBy,modifiedBy,sections"),
						},
						Then: mockserver.Response(http.StatusOK, estimateDetailEst1Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates/est-2"),
							mockcond.QueryParam("includes", "createdBy,modifiedBy,sections"),
						},
						Then: mockserver.ResponseString(http.StatusNotFound, `{"message":"Not Found"}`),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "est-1", "estimatenumber": "1042"},
						Raw:    map[string]any{"id": "est-1", "estimateNumber": "1042"},
					},
					{
						// Skipped detail: the row keeps the thin list stub.
						Fields: map[string]any{"id": "est-2"},
						Raw:    map[string]any{"id": "est-2", "isPrimary": false},
					},
				},
				Done: true,
			},
		},
		{
			// A transient failure (429/5xx) on one estimate's detail must not
			// sink the page — that row keeps its thin shape, the rest hydrate.
			Name: "Read estimates keeps the thin row when its detail returns 500",
			Input: common.ReadParams{
				ObjectName: "estimates",
				Fields:     connectors.Fields("id", "estimateNumber"),
				Opts:       ReadParamsOpts{HydrateEstimates: true},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates"),
						},
						Then: mockserver.Response(http.StatusOK, estimatesListResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates/est-1"),
							mockcond.QueryParam("includes", "createdBy,modifiedBy,sections"),
						},
						Then: mockserver.Response(http.StatusOK, estimateDetailEst1Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates/est-2"),
							mockcond.QueryParam("includes", "createdBy,modifiedBy,sections"),
						},
						Then: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"boom"}`),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "est-1", "estimatenumber": "1042"},
						Raw:    map[string]any{"id": "est-1", "estimateNumber": "1042"},
					},
					{
						// Skipped detail: the row keeps the thin list stub.
						Fields: map[string]any{"id": "est-2"},
						Raw:    map[string]any{"id": "est-2", "isPrimary": false},
					},
				},
				Done: true,
			},
		},
		{
			// Default (no Opts): rows pass through as list stubs and no
			// detail endpoint is called — the mock serves only the list, so
			// a stray hydration call would hit the 500 default and fail.
			Name: "Read estimates without opts returns stubs and skips hydration",
			Input: common.ReadParams{
				ObjectName: "estimates",
				Fields:     connectors.Fields("id", "isPrimary"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/api/v2/estimates"),
						},
						Then: mockserver.Response(http.StatusOK, estimatesListResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "est-1", "isprimary": true},
						Raw:    map[string]any{"id": "est-1", "isPrimary": true},
					},
					{
						Fields: map[string]any{"id": "est-2", "isprimary": false},
						Raw:    map[string]any{"id": "est-2", "isPrimary": false},
					},
				},
				Done: true,
			},
		},
		{
			Name: "Read acculynx/units-of-measure parses a bare array response",
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

// TestBareArrayObjectsResolveToEmptyResponseKey guards the generator's
// bareArrayObjects override. The AccuLynx spec declares a wrapper object for
// these three, but the live API answers with a top-level array, so their
// records key must be the empty bare-array convention. Without this, dropping
// an entry from the override map would only surface as a runtime
// "key not found" against the live API.
func TestBareArrayObjectsResolveToEmptyResponseKey(t *testing.T) {
	t.Parallel()

	connector, err := constructTestReadConnector("http://localhost")
	assert.NilError(t, err)

	tests := []struct {
		object string
		want   string
	}{
		{"acculynx/units-of-measure", ""},
		{"company-settings/job-file-settings/workflow-milestones", ""},
		{"company-settings/location-settings/account-types", ""},
		// Control: the spec describes countries the same way, but it genuinely wraps.
		{"acculynx/countries", "items"},
	}

	for _, tt := range tests {
		assert.Equal(t, connector.arrayFieldName(tt.object), tt.want, tt.object)
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

// jobsServer serves the empty custom-field definitions plus one /jobs request
// that must carry every given query param. Any other request gets a 500.
func jobsServer(query map[string]string, then http.HandlerFunc) *httptest.Server {
	jobs := mockcond.And{mockcond.MethodGET(), mockcond.Path("/api/v2/jobs")}
	for key, value := range query {
		jobs = append(jobs, mockcond.QueryParam(key, value))
	}

	return mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{If: jobs, Then: then},
			{
				If:   mockcond.Path("/api/v2/company-settings/custom-fields"),
				Then: mockserver.Response(http.StatusOK, customFieldDefinitionsEmptyResponse),
			},
		},
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
}

// jobsPage answers with one /jobs page of size jobs, at position start of a
// list holding count jobs. Every job is modified at first except the last one,
// which is modified at last.
func jobsPage(count, start, size int, first, last string) http.HandlerFunc {
	items := make([]string, size)
	for i := range items {
		modified := first
		if i == size-1 {
			modified = last
		}

		items[i] = fmt.Sprintf(`{"id": "job-%d", "modifiedDate": %q}`, start+i, modified)
	}

	return mockserver.ResponseString(http.StatusOK, fmt.Sprintf(
		`{"count": %d, "pageSize": %d, "pageStartIndex": %d, "items": [%s]}`,
		count, size, start, strings.Join(items, ",")))
}

func jobsURL(query string) common.NextPageToken {
	return common.NextPageToken(testconn.URLTestServer + "/api/v2/jobs?" + query)
}

// appointmentsServer serves the one-calendar list (cal-001) plus the given
// appointment window cases. A request for any other window falls through to
// the 500 default, so every window the connector asks for must be listed.
func appointmentsServer(windows ...mockserver.Case) *httptest.Server {
	cases := append([]mockserver.Case{{
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/v2/calendars"),
		},
		Then: mockserver.Response(http.StatusOK, calendarsListResponse),
	}}, windows...)

	return mockserver.Switch{
		Setup:   mockserver.ContentJSON(),
		Cases:   cases,
		Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
	}.Server()
}

// appointmentsWindowCase matches cal-001's appointments request for exactly one
// startDate/endDate window.
func appointmentsWindowCase(startDate, endDate string, then http.HandlerFunc) mockserver.Case {
	return mockserver.Case{
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/api/v2/calendars/cal-001/appointments"),
			mockcond.QueryParam("startDate", startDate),
			mockcond.QueryParam("endDate", endDate),
		},
		Then: then,
	}
}

// appointmentsPage answers with a single, final page holding the given items.
func appointmentsPage(items ...string) http.HandlerFunc {
	return mockserver.ResponseString(http.StatusOK, fmt.Sprintf(
		`{"count": %d, "pageSize": 25, "pageStartIndex": 0, "items": [%s]}`,
		len(items), strings.Join(items, ",")))
}

// TestAppointmentWindows pins the date windows read for calendars/appointments
// against a fixed clock: the defaults for a missing bound, and the split into
// windows of at most 89 days, since the API answers 400 for a range of 90 days
// or more (verified live).
func TestAppointmentWindows(t *testing.T) { //nolint:funlen
	t.Parallel()

	now := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)
	day := func(year int, month time.Month, d int) time.Time {
		return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
	}

	tests := []struct {
		name         string
		since, until time.Time
		want         []dateWindow
	}{
		{
			name:  "scheduled read: Since minutes ago, no Until, reads today to 60 days ahead",
			since: now.Add(-5 * time.Minute),
			want:  []dateWindow{{day(2026, 11, 4), day(2027, 1, 3)}},
		},
		{
			name: "no bounds: 30 days back to 60 days ahead, split at the 90-day limit",
			want: []dateWindow{
				{day(2026, 10, 5), day(2027, 1, 2)},
				{day(2027, 1, 3), day(2027, 1, 3)},
			},
		},
		{
			name:  "both bounds under the limit: one window, time of day dropped",
			since: time.Date(2026, 4, 1, 8, 30, 0, 0, time.UTC),
			until: time.Date(2026, 4, 30, 23, 59, 0, 0, time.UTC),
			want:  []dateWindow{{day(2026, 4, 1), day(2026, 4, 30)}},
		},
		{
			name:  "89-day range: one window",
			since: day(2026, 1, 1),
			until: day(2026, 3, 31),
			want:  []dateWindow{{day(2026, 1, 1), day(2026, 3, 31)}},
		},
		{
			name:  "90-day range: split",
			since: day(2026, 1, 1),
			until: day(2026, 4, 1),
			want: []dateWindow{
				{day(2026, 1, 1), day(2026, 3, 31)},
				{day(2026, 4, 1), day(2026, 4, 1)},
			},
		},
		{
			name:  "a year: contiguous windows of 89 days and the remainder",
			since: day(2026, 1, 1),
			until: day(2026, 12, 31),
			want: []dateWindow{
				{day(2026, 1, 1), day(2026, 3, 31)},
				{day(2026, 4, 1), day(2026, 6, 29)},
				{day(2026, 6, 30), day(2026, 9, 27)},
				{day(2026, 9, 28), day(2026, 12, 26)},
				{day(2026, 12, 27), day(2026, 12, 31)},
			},
		},
		{
			name:  "only Until, in the past: Since defaults to 30 days before Until",
			until: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
			want:  []dateWindow{{day(2026, 1, 30), day(2026, 3, 1)}},
		},
		{
			name:  "only Since, beyond the lookahead: Until defaults to 60 days after Since",
			since: day(2027, 6, 1),
			want:  []dateWindow{{day(2027, 6, 1), day(2027, 7, 31)}},
		},
		{
			name:  "non-UTC bounds are converted to UTC dates",
			since: time.Date(2026, 4, 1, 22, 0, 0, 0, time.FixedZone("UTC-5", -5*60*60)),
			until: day(2026, 4, 10),
			want:  []dateWindow{{day(2026, 4, 2), day(2026, 4, 10)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := appointmentWindows(common.ReadParams{Since: tt.since, Until: tt.until}, now)

			assert.Equal(t, len(got), len(tt.want), "window count")

			for i, w := range got {
				if !w.start.Equal(tt.want[i].start) || !w.end.Equal(tt.want[i].end) {
					t.Errorf("window %d = %s..%s, want %s..%s", i,
						w.start.Format(time.DateOnly), w.end.Format(time.DateOnly),
						tt.want[i].start.Format(time.DateOnly), tt.want[i].end.Format(time.DateOnly))
				}

				if span := w.end.Sub(w.start); span < 0 || span > maxAppointmentsWindowDays*24*time.Hour {
					t.Errorf("window %d spans %v, want 0..89 days", i, span)
				}
			}
		})
	}
}

// buildPaginationServer returns a server that answers every request with the
// given envelope plus numItems dummy records, so each test case shows the
// connector input, the envelope, and the expected pagination output in one
// place. Omit "count" from the envelope to simulate a response with no total.
func buildPaginationServer(t *testing.T, envelope map[string]int, numItems int) *httptest.Server {
	t.Helper()

	items := make([]map[string]any, numItems)
	for i := range items {
		items[i] = map[string]any{"id": "j" + strconv.Itoa(i), "jobNumber": strconv.Itoa(i)}
	}

	body := make(map[string]any, len(envelope)+1)
	for key, value := range envelope {
		body[key] = value
	}

	body["items"] = items

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal pagination response: %v", err)
	}

	return mockserver.Fixed{
		Setup:  mockserver.ContentJSON(),
		Always: mockserver.Response(http.StatusOK, payload),
	}.Server()
}

// TestReadPaginationTermination covers the exit conditions that keep Read from
// looping when AccuLynx misreports paging (CON-3512). Before the fix the
// connector's only exit was "a short page is the last page", so a server that
// answered every offset with a full first page paginated forever.
func TestReadPaginationTermination(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseRead{
		{
			// The regression guard: the server ignores pageStartIndex entirely,
			// answering every offset with a full page and a count it never
			// reaches. Read must still stop.
			Name: "Runaway server that ignores the offset terminates",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
				NextPage:   testconn.URLTestServer + "/api/v2/jobs?pageSize=25&pageStartIndex=25",
			},
			Server:     buildPaginationServer(t, map[string]int{"count": 163, "pageSize": 25, "pageStartIndex": 0}, 25),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				// The assigned sweep stops, and the read hands over to the
				// unassigned population exactly once.
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// Offset honoured: the first page must hand back a cursor.
			Name: "Count-based exit continues while records remain",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
			},
			Server:     buildPaginationServer(t, map[string]int{"count": 50, "pageSize": 25, "pageStartIndex": 0}, 25),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?includes=initialAppointment&pageSize=25&pageStartIndex=25",
				Done: false,
			},
		},
		{
			// Final page lands exactly on count with a full page of records, so
			// the short-page heuristic cannot fire — only the count check can.
			Name: "Count-based exit stops exactly at count on a full page",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
				NextPage:   testconn.URLTestServer + "/api/v2/jobs?pageSize=25&pageStartIndex=25",
			},
			Server:     buildPaginationServer(t, map[string]int{"count": 50, "pageSize": 25, "pageStartIndex": 25}, 25),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// We asked for offset 25; the envelope echoes 0. The server is
			// serving a page we have already seen, so stop rather than re-emit.
			Name: "Non-advancing echoed offset terminates",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
				NextPage:   testconn.URLTestServer + "/api/v2/jobs?pageSize=25&pageStartIndex=25",
			},
			Server:     buildPaginationServer(t, map[string]int{"count": 163, "pageSize": 25, "pageStartIndex": 0}, 25),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 25,
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&pageSize=25&pageStartIndex=0",
				Done: false,
			},
		},
		{
			// A count-reporting account is bounded no matter how deep the sweep
			// goes, so an arbitrarily large offset must keep paging normally.
			Name: "Deep offset keeps paging when count is reported",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
				NextPage:   testconn.URLTestServer + "/api/v2/jobs?pageSize=25&pageStartIndex=10000",
			},
			Server:     buildPaginationServer(t, map[string]int{"count": 50000, "pageSize": 25, "pageStartIndex": 10000}, 25),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     25,
				NextPage: testconn.URLTestServer + "/api/v2/jobs?pageSize=25&pageStartIndex=10025",
				Done:     false,
			},
		},
		{
			// Full page, offset honoured, no count: nothing can bound the sweep,
			// so the read must error immediately and name the object.
			Name: "Missing count on a full page errors immediately",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
			},
			Server:       buildPaginationServer(t, map[string]int{"pageSize": 25, "pageStartIndex": 0}, 25),
			ExpectedErrs: []error{errMissingCount},
		},
		{
			// No count, but the page is short — that is an unambiguous last page,
			// so the read completes rather than erroring.
			Name: "Missing count on a short page still terminates cleanly",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("id"),
				PageSize:   25,
			},
			Server:     buildPaginationServer(t, map[string]int{"pageSize": 25, "pageStartIndex": 0}, 10),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 10,
				NextPage: testconn.URLTestServer +
					"/api/v2/jobs?assignment=unassigned&includes=initialAppointment&pageSize=25&pageStartIndex=0",
				Done: false,
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
