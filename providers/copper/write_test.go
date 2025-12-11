package copper

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseUnprocessableEntity := testutils.DataFromFile(t, "write/unprocessable-entity.json")
	responseCompanies := testutils.DataFromFile(t, "write/companies/new.json")
	responseProjects := testutils.DataFromFile(t, "write/projects/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error unprocessable entity",
			Input: common.WriteParams{ObjectName: "companies", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnprocessableEntity, responseUnprocessableEntity),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Invalid input: Validation errors: Name: can't be blank"),
			},
		},
		{
			Name:  "Create company via POST",
			Input: common.WriteParams{ObjectName: "companies", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/developer_api/v1/companies"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "73615382",
				Errors:   nil,
				Data: map[string]any{
					"name":          "Demo Company",
					"date_created":  float64(1754503301),
					"date_modified": float64(1754503301),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update company via PUT",
			Input: common.WriteParams{ObjectName: "companies", RecordId: "73615382", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/developer_api/v1/companies/73615382"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "73615382",
				Errors:   nil,
				Data: map[string]any{
					"name":          "Demo Company",
					"date_created":  float64(1754503301),
					"date_modified": float64(1754503301),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create project via POST",
			Input: common.WriteParams{ObjectName: "projects", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/developer_api/v1/projects"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseProjects),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1621193",
				Errors:   nil,
				Data: map[string]any{
					"name":   "Great project",
					"status": "Open",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update project via PUT",
			Input: common.WriteParams{ObjectName: "projects", RecordData: "dummy", RecordId: "1621193"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/developer_api/v1/projects/1621193"),
					mockcond.Header(http.Header{"X-PW-Application": []string{"developer_api"}}),
					mockcond.Header(http.Header{"X-PW-UserEmail": []string{"john@test.com"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseProjects),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1621193",
				Errors:   nil,
				Data: map[string]any{
					"name":   "Great project",
					"status": "Open",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
