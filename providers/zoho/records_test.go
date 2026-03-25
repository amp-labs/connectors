package zoho

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

type GetRecordsByIdsInput struct {
	ObjectName   string
	Ids          []string
	Fields       []string
	Associations []string
}

func TestGetRecordsByIds(t *testing.T) {
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")

	tests := []testroutines.TestCase[GetRecordsByIdsInput, []common.ReadResultRow]{
		{
			Name: "Empty record IDs returns error",
			Input: GetRecordsByIdsInput{
				ObjectName: "contacts",
				Fields:     []string{"Full_Name"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{errMissingParams},
		},
		{
			Name: "Successfully fetch contacts by IDs",
			Input: GetRecordsByIdsInput{
				ObjectName: "contacts",
				Fields:     []string{"Full_Name"},
				Ids:        []string{"6493490000001291001", "6493490000001291002"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/crm/v6/contacts"),
					mockcond.QueryParam("ids", "6493490000001291001,6493490000001291002"),
				},
				Then: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Expected: []common.ReadResultRow{
				{
					Id: "6493490000001291001",
					Fields: map[string]any{
						"id":        "6493490000001291001",
						"full_name": "Ryan Dahl2",
					},
					Raw: map[string]any{
						"Full_Name":    "Ryan Dahl2",
						"id":           "6493490000001291001",
						"Assistant":    nil,
						"Created_Time": "2024-12-20T10:09:52+03:00",
						"Created_By": map[string]any{
							"name":  "Joseph Karage",
							"id":    "6493490000000486001",
							"email": "josephkarage@gmail.com",
						},
					},
				},
				{
					Id: "6493490000001291002",
					Fields: map[string]any{
						"id":        "6493490000001291002",
						"full_name": "Jane Smith",
					},
					Raw: map[string]any{
						"Full_Name": "Jane Smith",
						"id":        "6493490000001291002",
						"Created_By": map[string]any{
							"name":  "Joseph Karage",
							"id":    "6493490000000486001",
							"email": "josephkarage@gmail.com",
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

			t.Cleanup(func() {
				tt.Close()
			})

			conn, err := constructTestConnector(tt.Server.URL)
			if err != nil {
				t.Fatalf("failed to construct test connector: %v", err)
			}

			result, err := conn.GetRecordsByIds(t.Context(), tt.Input.ObjectName, tt.Input.Ids, tt.Input.Fields, tt.Input.Associations)

			tt.Validate(t, err, result)
		})
	}
}
