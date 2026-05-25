package gusto

import (
	_ "embed"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

//go:embed test/read/employees-first-page.json
var employeesFirstPageResponse []byte

//go:embed test/read/employees-last-page.json
var employeesLastPageResponse []byte

//go:embed test/read/companies.json
var companiesResponse []byte

//go:embed test/read/jobs-emp-001.json
var jobsEmp001Response []byte

//go:embed test/read/jobs-emp-002.json
var jobsEmp002Response []byte

const testCompanyID = "test-company-uuid"

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Object must be supported",
			Input: common.ReadParams{
				ObjectName: "compensations",
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Company-scoped reads require companyId metadata",
			Input: common.ReadParams{
				ObjectName: "employees",
				Fields:     connectors.Fields("uuid"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingCompanyID},
		},
		{
			Name: "Employee-scoped reads require companyId metadata",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("uuid"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingCompanyID},
		},
		{
			Name: "Read employees with full page returns next page",
			Input: common.ReadParams{
				ObjectName: "employees",
				Fields:     connectors.Fields("uuid", "first_name", "last_name"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/employees"),
							mockcond.QueryParam("per", "2"),
							mockcond.QueryParam("page", "1"),
						},
						Then: mockserver.Response(http.StatusOK, employeesFirstPageResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"uuid":       "emp_001",
						"first_name": "Alice",
						"last_name":  "Anderson",
					},
					// Raw must include "email" even though it wasn't requested in Fields,
					// proving the raw response is preserved as-is.
					Raw: map[string]any{
						"uuid":       "emp_001",
						"first_name": "Alice",
						"last_name":  "Anderson",
						"email":      "alice@example.com",
					},
				}, {
					Fields: map[string]any{
						"uuid":       "emp_002",
						"first_name": "Bob",
						"last_name":  "Brown",
					},
					Raw: map[string]any{
						"uuid":       "emp_002",
						"first_name": "Bob",
						"last_name":  "Brown",
						"email":      "bob@example.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v1/companies/" + testCompanyID + "/employees?page=2&per=2",
				Done:     false,
			},
		},
		{
			Name: "Read employees partial page signals Done",
			Input: common.ReadParams{
				ObjectName: "employees",
				Fields:     connectors.Fields("uuid"),
				PageSize:   2,
				NextPage:   testroutines.URLTestServer + "/v1/companies/" + testCompanyID + "/employees?page=2&per=2",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/employees"),
							mockcond.QueryParam("per", "2"),
							mockcond.QueryParam("page", "2"),
						},
						Then: mockserver.Response(http.StatusOK, employeesLastPageResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read companies does not require companyId substitution",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("uuid", "name"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/companies"),
						},
						Then: mockserver.Response(http.StatusOK, companiesResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"uuid": "comp_001",
						"name": "Ampersand Test Co",
					},
					Raw: map[string]any{
						"uuid": "comp_001",
						"name": "Ampersand Test Co",
						"ein":  "12-3456789",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Empty response returns Done",
			Input: common.ReadParams{
				ObjectName: "employees",
				Fields:     connectors.Fields("uuid"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.MethodGET(),
						Then: mockserver.Response(http.StatusOK, []byte(`[]`)),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read jobs fans out per employee and returns flattened records",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("uuid", "title", "employee_uuid"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/employees"),
							mockcond.QueryParam("per", "2"),
							mockcond.QueryParam("page", "1"),
						},
						Then: mockserver.Response(http.StatusOK, employeesFirstPageResponse),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/employees/emp_001/jobs"),
						},
						Then: mockserver.Response(http.StatusOK, jobsEmp001Response),
					},
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/employees/emp_002/jobs"),
						},
						Then: mockserver.Response(http.StatusOK, jobsEmp002Response),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				// 2 jobs from emp_001 + 1 job from emp_002 = 3 total
				Rows: 3,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"uuid":          "job_001",
						"title":         "Software Engineer",
						"employee_uuid": "emp_001",
					},
					// Raw must include "rate", "payment_unit", "primary" even though
					// they weren't requested — proves the child-fetch response is
					// preserved intact through the fan-out.
					Raw: map[string]any{
						"uuid":          "job_001",
						"employee_uuid": "emp_001",
						"title":         "Software Engineer",
						"rate":          "80000.00",
						"payment_unit":  "Year",
						"primary":       true,
					},
				}},
				// Employee list had exactly 2 rows (== per=2), so there may be more employees.
				NextPage: testroutines.URLTestServer + "/v1/companies/" + testCompanyID + "/employees?page=2&per=2",
				Done:     false,
			},
		},
		{
			Name: "Read jobs with empty employee list returns Done",
			Input: common.ReadParams{
				ObjectName: "jobs",
				Fields:     connectors.Fields("uuid"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodGET(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/employees"),
						},
						Then: mockserver.Response(http.StatusOK, []byte(`[]`)),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				noCompanyIDCases := map[string]bool{
					"Company-scoped reads require companyId metadata":  true,
					"Employee-scoped reads require companyId metadata": true,
				}

				if noCompanyIDCases[tt.Name] {
					return constructTestReadConnector(tt.Server.URL, "")
				}

				return constructTestReadConnector(tt.Server.URL, testCompanyID)
			})
		})
	}
}

func constructTestReadConnector(baseURL, companyID string) (*Connector, error) {
	meta := map[string]string{}
	if companyID != "" {
		meta[metadataKeyCompanyID] = companyID
	}

	conn, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
		Metadata:            meta,
	})
	if err != nil {
		return nil, err
	}

	conn.SetUnitTestBaseURL(baseURL)

	return conn, nil
}
