package procore

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createProjectResponse := testutils.DataFromFile(t, "create_project.json")
	updateProjectResponse := testutils.DataFromFile(t, "update_project.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Write object must be supported",
			Input: common.WriteParams{
				ObjectName: "submittal_statuses",
				RecordData: map[string]any{"name": "x"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create project posts to company-scoped endpoint",
			Input: common.WriteParams{
				ObjectName: "projects",
				RecordData: map[string]any{
					"project": map[string]any{
						"name":           "Test Project Charlie",
						"project_number": "PN-003",
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects"),
					mockcond.Header(http.Header{"Procore-Company-Id": []string{testCompanyID}}),
				},
				Then: mockserver.Response(http.StatusCreated, createProjectResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2783407",
				Data: map[string]any{
					"id":             float64(2783407),
					"name":           "Test Project Charlie",
					"project_number": "PN-003",
					"active":         true,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update project patches to /{id}",
			Input: common.WriteParams{
				ObjectName: "projects",
				RecordId:   "2783405",
				RecordData: map[string]any{
					"project": map[string]any{"name": "Test Project Alpha (renamed)"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects/2783405"),
				},
				Then: mockserver.Response(http.StatusOK, updateProjectResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2783405",
				Data: map[string]any{
					"id":   float64(2783405),
					"name": "Test Project Alpha (renamed)",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create vendor uses top-level path with company_id query param",
			Input: common.WriteParams{
				ObjectName: "vendors",
				RecordData: map[string]any{"vendor": map[string]any{"name": "Acme Subs"}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/v1.0/vendors"),
					mockcond.QueryParam("company_id", testCompanyID),
				},
				Then: mockserver.Response(http.StatusCreated, createProjectResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2783407",
				Data:     map[string]any{"id": float64(2783407)},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update vendor patches /{id} keeping company_id query param",
			Input: common.WriteParams{
				ObjectName: "vendors",
				RecordId:   "9001",
				RecordData: map[string]any{"vendor": map[string]any{"name": "Acme Subs LLC"}},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/rest/v1.0/vendors/9001"),
					mockcond.QueryParam("company_id", testCompanyID),
				},
				Then: mockserver.Response(http.StatusOK, updateProjectResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2783405",
				Data:     map[string]any{"id": float64(2783405)},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
