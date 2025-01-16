package zohocrm

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts.json")
	callsResponse := testutils.DataFromFile(t, "calls.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupportedread.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "arsenal", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrCaller,
				errors.New(string(unsupportedResponse)), //nolint:err113
			},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "calls", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, callsResponse),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Contacts",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("Assistant", "Created_By", "Full_Name", "id", "Created_Time"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"assistant": nil,
						"created_by": map[string]any{
							"email": "josephkarage@gmail.com",
							"id":    "6493490000000486001",
							"name":  "Joseph Karage",
						},
						"created_time": "2024-12-20T10:09:52+03:00",
						"full_name":    "Ryan Dahl2",
						"id":           "6493490000001291001",
					},
					Raw: map[string]any{
						"Assistant": nil,
						"Created_By": map[string]any{
							"email": "josephkarage@gmail.com",
							"id":    "6493490000000486001",
							"name":  "Joseph Karage",
						},
						"Created_Time": "2024-12-20T10:09:52+03:00",
						"Full_Name":    "Ryan Dahl2",
						"id":           "6493490000001291001",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
