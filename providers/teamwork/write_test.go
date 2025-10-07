package teamwork

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

	responseConflictError := testutils.DataFromFile(t, "write/conflict.json")
	responseCompanies := testutils.DataFromFile(t, "write/companies/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error invalid payload",
			Input: common.WriteParams{ObjectName: "companies", RecordData: map[string]any{}},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusConflict, responseConflictError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("conflict; Company name already in use"), // nolint:goerr113
			},
		},
		{
			Name: "Create company via POST",
			Input: common.WriteParams{
				ObjectName: "companies",
				RecordData: map[string]any{
					"city": "Vancouver",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/projects/api/v3/companies.json"),
					mockcond.Body(`{"company":{"city":"Vancouver"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1412996",
				Errors:   nil,
				Data: map[string]any{
					"name": "Nike",
					"city": "Vancouver",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update task via PUT",
			Input: common.WriteParams{
				ObjectName: "companies",
				RecordId:   "1412996",
				RecordData: map[string]any{
					"city": "Vancouver",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/projects/api/v3/companies/1412996.json"),
					mockcond.Body(`{"company":{"city":"Vancouver"}}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1412996",
				Errors:   nil,
				Data: map[string]any{
					"name": "Nike",
					"city": "Vancouver",
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
