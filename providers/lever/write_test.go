// nolint:dupl
package lever

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	requisitionFieldResponse := testutils.DataFromFile(t, "write_requisition_field.json")
	requisitionsResponse := testutils.DataFromFile(t, "write_requisitions.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the requisitions",
			Input: common.WriteParams{ObjectName: "requisitions", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/requisitions"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, requisitionsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "33b3cdf8-baef-4b17-b9d5-04c3b7e5eeb2",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"id":              "33b3cdf8-baef-4b17-b9d5-04c3b7e5eeb2",
						"requisitionCode": "ENG-9",
						"name":            "Junior Software Engineer, Platform",
						"backfill":        false,
						"createdAt":       float64(1452306348935),
						"creator":         "d68cdaeb-fc0f-462a-b9a5-37bdbdeb0c68",
						"headcountHired":  float64(0),
						"headcountTotal":  float64(10),
						"owner":           "f1f9035d-38f4-4e18-82ae-f2eac3e2592a",
						"status":          "open",
						"hiringManager":   "d68cdaeb-fc0f-462a-b9a5-37bdbdeb0c68",
						"approval": map[string]any{
							"steps": []any{},
						},
						"compensationBand": map[string]any{
							"currency": "USD",
							"interval": "per-year-salary",
							"min":      float64(100000),
							"max":      float64(130000),
						},
						"employmentStatus": "full-time",
						"location":         "New York",
						"internalNotes":    "College grad hire -- very little flexibility on salary",
						"postings":         []any{},
						"team":             "Product Engineering",
						"offerIds":         []any{},
						"customFields": map[string]any{
							"cost_center": map[string]any{
								"city":        "New York",
								"campus_code": "9",
							},
							"target_hire_date": float64(1452067200000),
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update requisitions as PUT",
			Input: common.WriteParams{
				ObjectName: "requisitions",
				RecordId:   "33b3cdf8-baef-4b17-b9d5-04c3b7e5eeb2",
				RecordData: "dummy",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/requisitions/33b3cdf8-baef-4b17-b9d5-04c3b7e5eeb2"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, requisitionsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "33b3cdf8-baef-4b17-b9d5-04c3b7e5eeb2",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"id":              "33b3cdf8-baef-4b17-b9d5-04c3b7e5eeb2",
						"requisitionCode": "ENG-9",
						"name":            "Junior Software Engineer, Platform",
						"backfill":        false,
						"createdAt":       float64(1452306348935),
						"creator":         "d68cdaeb-fc0f-462a-b9a5-37bdbdeb0c68",
						"headcountHired":  float64(0),
						"headcountTotal":  float64(10),
						"owner":           "f1f9035d-38f4-4e18-82ae-f2eac3e2592a",
						"status":          "open",
						"hiringManager":   "d68cdaeb-fc0f-462a-b9a5-37bdbdeb0c68",
						"approval": map[string]any{
							"steps": []any{},
						},
						"compensationBand": map[string]any{
							"currency": "USD",
							"interval": "per-year-salary",
							"min":      float64(100000),
							"max":      float64(130000),
						},
						"employmentStatus": "full-time",
						"location":         "New York",
						"internalNotes":    "College grad hire -- very little flexibility on salary",
						"postings":         []any{},
						"team":             "Product Engineering",
						"offerIds":         []any{},
						"customFields": map[string]any{
							"cost_center": map[string]any{
								"city":        "New York",
								"campus_code": "9",
							},
							"target_hire_date": float64(1452067200000),
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating the requisition_fields",
			Input: common.WriteParams{ObjectName: "requisition_fields", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/requisition_fields"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, requisitionFieldResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cost_center",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"id":         "cost_center",
						"text":       "Cost center",
						"type":       "object",
						"isRequired": true,
						"subfields": []any{
							map[string]any{
								"id":   "city",
								"text": "City",
								"type": "text",
							},
							map[string]any{
								"id":   "campus_code",
								"text": "Campus code",
								"type": "number",
							},
							map[string]any{
								"id":   "department",
								"text": "Department",
								"type": "text",
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Updating the requisition_fields",
			Input: common.WriteParams{ObjectName: "requisition_fields", RecordId: "cost_center", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/requisition_fields/cost_center"),
					mockcond.MethodPUT(),
				},
				Then: mockserver.Response(http.StatusOK, requisitionFieldResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cost_center",
				Errors:   nil,
				Data: map[string]any{
					"data": map[string]any{
						"id":         "cost_center",
						"text":       "Cost center",
						"type":       "object",
						"isRequired": true,
						"subfields": []any{
							map[string]any{
								"id":   "city",
								"text": "City",
								"type": "text",
							},
							map[string]any{
								"id":   "campus_code",
								"text": "Campus code",
								"type": "number",
							},
							map[string]any{
								"id":   "department",
								"text": "Department",
								"type": "text",
							},
						},
					},
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
