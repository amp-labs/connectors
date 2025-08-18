package xero

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

	createContactGroupResponse := testutils.DataFromFile(t, "create-contactGroups.json")
	updateContactGroupResponse := testutils.DataFromFile(t, "update-contactGroups.json")

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
			Name: "Successfully creation of an contactGroups",
			Input: common.WriteParams{ObjectName: "contactGroups", RecordData: map[string]any{
				"name": "Bert Olson",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, createContactGroupResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "055712c7-0fcf-4ba2-a900-1200f10cfe28",
				Data: map[string]any{
					"ContactGroups": []any{
						map[string]any{
							"ContactGroupID":      "efd27a1e-a1f5-4cbf-8439-b0f912af709c",
							"HasValidationErrors": false,
							"Name":                "Bert Olson",
							"Status":              "ACTIVE",
						},
					},
					"DateTimeUTC":  "/Date(1755500880867)/",
					"Id":           "055712c7-0fcf-4ba2-a900-1200f10cfe28",
					"ProviderName": "Ampersand test",
					"Status":       "OK",
				},
			},
			ExpectedErrs: nil,
		},

		{
			Name: "Successfully updated ContactGroups",
			Input: common.WriteParams{ObjectName: "contactGroups",
				RecordId: "055712c7-0fcf-4ba2-a900-1200f10cfe28",
				RecordData: map[string]any{
					"Name": "Eusebio Damore",
				}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, updateContactGroupResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "055712c7-0fcf-4ba2-a900-1200f10cfe28",
				Data: map[string]any{
					"ContactGroups": []any{
						map[string]any{
							"ContactGroupID":      "efd27a1e-a1f5-4cbf-8439-b0f912af709c",
							"HasValidationErrors": false,
							"Name":                "Eusebio Damore",
							"Status":              "ACTIVE",
						},
					},
					"DateTimeUTC":  "/Date(1755500881326)/",
					"Id":           "055712c7-0fcf-4ba2-a900-1200f10cfe28",
					"ProviderName": "Ampersand test",
					"Status":       "OK",
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
