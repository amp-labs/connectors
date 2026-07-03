package gusto

import (
	_ "embed"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

//go:embed test/write/employee-create.json
var employeeCreateResponse []byte

//go:embed test/write/employee-update.json
var employeeUpdateResponse []byte

//go:embed test/write/job-create.json
var jobCreateResponse []byte

//go:embed test/write/compensation-create.json
var compensationCreateResponse []byte

//go:embed test/write/location-update.json
var locationUpdateResponse []byte

//go:embed test/write/earning-type-update.json
var earningTypeUpdateResponse []byte

func TestWrite(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Object must be supported",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordData: map[string]any{"name": "test"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Update on companies returns the updated record",
			// companies is updateOnly — no create allowed.
			Input: common.WriteParams{
				ObjectName: "companies",
				RecordData: map[string]any{"contractor_only": false},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Update on admins is not supported",
			// admins is createOnly — Gusto does not expose PUT.
			Input: common.WriteParams{
				ObjectName: "admins",
				RecordId:   "adm_001",
				RecordData: map[string]any{"first_name": "x", "version": "v1"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create employee succeeds — companyID injected from connector metadata",
			Input: common.WriteParams{
				ObjectName: "employees",
				RecordData: map[string]any{
					"first_name": "Alice",
					"last_name":  "Anderson",
					"email":      "alice@example.com",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/employees"),
						},
						Then: mockserver.Response(http.StatusCreated, employeeCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "emp_001",
				Data: map[string]any{
					"uuid":       "emp_001",
					"first_name": "Alice",
				},
			},
		},
		{
			Name: "Update employee succeeds — top-level URL by uuid",
			// PUT requires `version` field in RecordData; we don't synthesize it.
			Input: common.WriteParams{
				ObjectName: "employees",
				RecordId:   "emp_001",
				RecordData: map[string]any{
					"first_name": "Alicia",
					"version":    "v1",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPUT(),
							mockcond.Path("/v1/employees/emp_001"),
						},
						Then: mockserver.Response(http.StatusOK, employeeUpdateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "emp_001",
				Data: map[string]any{
					"uuid":       "emp_001",
					"first_name": "Alicia",
					"version":    "v2",
				},
			},
		},
		{
			Name: "Update location succeeds — top-level URL by uuid",
			Input: common.WriteParams{
				ObjectName: "locations",
				RecordId:   "loc_001",
				RecordData: map[string]any{
					"city":    "San Francisco",
					"version": "v1",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPUT(),
							mockcond.Path("/v1/locations/loc_001"),
						},
						Then: mockserver.Response(http.StatusOK, locationUpdateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "loc_001",
				Data: map[string]any{
					"uuid":    "loc_001",
					"version": "v2",
				},
			},
		},
		{
			Name: "Update earning_type uses company-scoped URL — PUT /v1/companies/{cid}/earning_types/{uuid}",
			// earning_types, pay_schedules, and payrolls have a nested PUT URL
			// per Gusto's docs (sitemap entry put-v1-companies-company_id-earning_types-earning_type_uuid).
			// All other top-level updates flatten to PUT /v1/{object}/{uuid}.
			Input: common.WriteParams{
				ObjectName: "earning_types",
				RecordId:   "et_001",
				RecordData: map[string]any{
					"name":    "Bonus Updated",
					"version": "v1",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPUT(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/earning_types/et_001"),
						},
						Then: mockserver.Response(http.StatusOK, earningTypeUpdateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "et_001",
				Data: map[string]any{
					"uuid":    "et_001",
					"name":    "Bonus Updated",
					"version": "v2",
				},
			},
		},
		{
			Name: "Update earning_type without companyID returns ErrMissingCompanyID",
			// Same nested-URL path requires companyID metadata. Sanity check that
			// we error early rather than constructing a malformed URL.
			Input: common.WriteParams{
				ObjectName: "earning_types",
				RecordId:   "et_001",
				RecordData: map[string]any{"name": "x", "version": "v1"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingCompanyID},
		},
		{
			Name: "Create job — employee_id extracted from RecordData and stripped from body",
			Input: common.WriteParams{
				ObjectName: "jobs",
				RecordData: map[string]any{
					"employee_id": "emp_001",
					"title":       "Software Engineer",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v1/employees/emp_001/jobs"),
						},
						Then: mockserver.Response(http.StatusCreated, jobCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "job_001",
				Data: map[string]any{
					"uuid": "job_001",
				},
			},
		},
		{
			Name: "Create job missing employee_id returns errMissingParentID",
			Input: common.WriteParams{
				ObjectName: "jobs",
				RecordData: map[string]any{"title": "Software Engineer"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParentID},
		},
		{
			Name: "Create compensation — job_id extracted from RecordData",
			Input: common.WriteParams{
				ObjectName: "compensations",
				RecordData: map[string]any{
					"job_id":       "job_001",
					"rate":         "100.00",
					"payment_unit": "Hour",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodPOST(),
							mockcond.Path("/v1/jobs/job_001/compensations"),
						},
						Then: mockserver.Response(http.StatusCreated, compensationCreateResponse),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "comp_001",
				Data: map[string]any{
					"uuid": "comp_001",
				},
			},
		},
		{
			Name: "Create employee without companyID returns ErrMissingCompanyID",
			Input: common.WriteParams{
				ObjectName: "employees",
				RecordData: map[string]any{"first_name": "Alice"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingCompanyID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				// Cases that test ErrMissingCompanyID explicitly omit the metadata.
				switch tt.Name {
				case "Create employee without companyID returns ErrMissingCompanyID",
					"Update earning_type without companyID returns ErrMissingCompanyID":
					return constructTestWriteConnector(tt.Server.URL, "")
				}

				return constructTestWriteConnector(tt.Server.URL, testCompanyID)
			})
		})
	}
}

func constructTestWriteConnector(baseURL, companyID string) (*Connector, error) {
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
