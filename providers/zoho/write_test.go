package zoho

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	unsupportedResponse := testutils.DataFromFile(t, "unsupportedread.json")
	leadsWriteResponse := testutils.DataFromFile(t, "leads-write.json")
	updateContactsResponse := testutils.DataFromFile(t, "updatecontact.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "RecordData must be an object or an array of objects",
			Input:        common.WriteParams{ObjectName: "Leads", RecordData: "hahaha"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrBadRequest},
		},
		{
			Name: "Unsupported object",
			Input: common.WriteParams{ObjectName: "arsenal", RecordData: map[string]any{
				"key": "value",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusBadRequest, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				testutils.StringError(string(unsupportedResponse)),
			},
		},
		{
			Name: "Successfully Create a Lead",
			Input: common.WriteParams{ObjectName: "leads", RecordData: map[string]any{
				"Lead_Source": "Employee Referral",
				"Company":     "Ampersand",
				"Last_Name":   "Daniel",
				"First_Name":  "Alexia",
				"Email":       "a.daly@zylker.com",
				"State":       "Texas",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, leadsWriteResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6493490000001504003",
				Data: map[string]any{
					"code": "SUCCESS",
					"details": map[string]any{
						"Modified_Time": "2025-01-10T13:09:14+03:00",
						"Modified_By":   map[string]any{"name": "Joseph Karage", "id": "6493490000000486001"},
						"Created_Time":  "2025-01-10T13:09:14+03:00",
						"id":            "6493490000001504003",
						"Created_By":    map[string]any{"name": "Joseph Karage", "id": "6493490000000486001"},
					},
					"message": "record added",
					"status":  "success",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update a contact",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "7e5209b8-bd4e-41d9-bbcd-2f9bab7d4030",
				RecordData: map[string]any{
					"Name": "John Snow",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, updateContactsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "6493490000001291001",
				Data: map[string]any{
					"code": "SUCCESS",
					"details": map[string]any{
						"Modified_Time": "2025-01-10T13:22:08+03:00",
						"Modified_By":   map[string]any{"name": "Joseph Karage", "id": "6493490000000486001"},
						"Created_Time":  "2024-12-20T10:09:52+03:00",
						"id":            "6493490000001291001",
						"Created_By":    map[string]any{"name": "Joseph Karage", "id": "6493490000000486001"},
					},
					"message": "record updated",
					"status":  "success",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
