package housecallpro

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

func TestWrite(t *testing.T) {
	t.Parallel()

	responseCustomerCreate := testutils.DataFromFile(t, "write-customer-create.json")
	responseCustomerUpdate := testutils.DataFromFile(t, "write-customer-update.json")
	responseJobTypeCreate := testutils.DataFromFile(t, "write-job-type-create.json")
	responseJobTypeUpdate := testutils.DataFromFile(t, "write-job-type-update.json")
	responseMaterialCreate := testutils.DataFromFile(t, "write-material-create.json")
	responseMaterialUpdate := testutils.DataFromFile(t, "write-material-update.json")

	tests := []testroutines.Write{
		{
			Name: "Create customer successfully",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{
					"first_name": "Levi",
					"last_name":  "Ackerman",
					"email":      "levi.ackerman@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/customers"),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cust_9f3a7c2b1d",
				Data: map[string]any{
					"id":         "cust_9f3a7c2b1d",
					"first_name": "Levi",
					"last_name":  "Ackerman",
					"email":      "levi.ackerman@example.com",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update customer successfully",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "cust_9f3a7c2b1d",
				RecordData: map[string]any{
					"last_name": "Ackerman Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/customers/cust_9f3a7c2b1d"),
					mockcond.Body(`{"last_name":"Ackerman Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cust_9f3a7c2b1d",
				Data: map[string]any{
					"last_name": "Ackerman Updated",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create job type successfully",
			Input: common.WriteParams{
				ObjectName: "job_fields/job_types",
				RecordData: map[string]any{
					"name": "Plumbing",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/job_fields/job_types"),
				},
				Then: mockserver.Response(http.StatusOK, responseJobTypeCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "job_type_123",
				Data: map[string]any{
					"id":   "job_type_123",
					"name": "Plumbing",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update job type successfully",
			Input: common.WriteParams{
				ObjectName: "job_fields/job_types",
				RecordId:   "job_type_123",
				RecordData: map[string]any{
					"name": "Plumbing Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/job_fields/job_types/job_type_123"),
					mockcond.Body(`{"name":"Plumbing Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseJobTypeUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "job_type_123",
				Data: map[string]any{
					"id":   "job_type_123",
					"name": "Plumbing Updated",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create price book material successfully",
			Input: common.WriteParams{
				ObjectName: "price_book/materials",
				RecordData: map[string]any{
					"name":                   "Premium Wall Paint",
					"description":            "High-quality interior wall paint with long-lasting finish",
					"material_category_uuid": "pbmatcat_7f3a21",
					"unit_of_measure":        "gallon",
					"taxable":                true,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/api/price_book/materials"),
				},
				Then: mockserver.Response(http.StatusOK, responseMaterialCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "mat_8c92ab4f1e",
				Data: map[string]any{
					"object": "material",
					"uuid":   "mat_8c92ab4f1e",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update price book material successfully",
			Input: common.WriteParams{
				ObjectName: "price_book/materials",
				RecordId:   "mat_8c92ab4f1e",
				RecordData: map[string]any{
					"name": "Premium Wall Paint Updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPut),
					mockcond.Path("/api/price_book/materials/mat_8c92ab4f1e"),
					mockcond.Body(`{"name":"Premium Wall Paint Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseMaterialUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "mat_8c92ab4f1e",
				Data: map[string]any{
					"name": "Premium Wall Paint Updated",
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
