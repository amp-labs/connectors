package gusto

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete record ID must be included",
			Input:        common.DeleteParams{ObjectName: "jobs"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name: "Unsupported object returns ErrOperationNotSupportedForObject",
			// locations has no DELETE per Gusto's docs; the connector rejects.
			Input:        common.DeleteParams{ObjectName: "locations", RecordId: "loc_001"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Top-level DELETE on jobs hits /v1/jobs/{uuid}",
			Input: common.DeleteParams{ObjectName: "jobs", RecordId: "job_001"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodDELETE(),
							mockcond.Path("/v1/jobs/job_001"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Top-level DELETE on home_addresses hits /v1/home_addresses/{uuid}",
			Input: common.DeleteParams{ObjectName: "home_addresses", RecordId: "addr_001"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodDELETE(),
							mockcond.Path("/v1/home_addresses/addr_001"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name: "Company-scoped DELETE on earning_types hits /v1/companies/{cid}/earning_types/{uuid}",
			// earning_types delete is nested under company per Gusto's docs
			// (delete-v1-companies-company_id-earning_types-earning_type_uuid).
			Input: common.DeleteParams{ObjectName: "earning_types", RecordId: "et_001"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.MethodDELETE(),
							mockcond.Path("/v1/companies/" + testCompanyID + "/earning_types/et_001"),
						},
						Then: mockserver.Response(http.StatusNoContent, nil),
					},
				},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected"}`),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:         "DELETE earning_types without companyID returns ErrMissingCompanyID",
			Input:        common.DeleteParams{ObjectName: "earning_types", RecordId: "et_001"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrMissingCompanyID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				if tt.Name == "DELETE earning_types without companyID returns ErrMissingCompanyID" {
					return constructTestWriteConnector(tt.Server.URL, "")
				}

				return constructTestWriteConnector(tt.Server.URL, testCompanyID)
			})
		})
	}
}
