package zohocrm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	dealResponse := testutils.DataFromFile(t, "deals.json")
	arsenalResponse := testutils.DataFromFile(t, "arsenal.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"deals", "arsenal"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.QueryParam("module", "Deals"),
					Then: mockserver.Response(http.StatusOK, dealResponse),
				}, {
					If:   mockcond.QueryParam("module", "Arsenal"),
					Then: mockserver.Response(http.StatusBadRequest, arsenalResponse),
				}}}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Deals": {
						DisplayName: "Deals",
						FieldsMap: map[string]string{
							"Account_Name":           "Account_Name",
							"Amount":                 "Amount",
							"Campaign_Source":        "Campaign_Source",
							"Change_Log_Time__s":     "Change_Log_Time__s",
							"Closing_Date":           "Closing_Date",
							"Contact_Name":           "Contact_Name",
							"Created_By":             "Created_By",
							"Created_Time":           "Created_Time",
							"Deal_Name":              "Deal_Name",
							"Description":            "Description",
							"Expected_Revenue":       "Expected_Revenue",
							"Last_Activity_Time":     "Last_Activity_Time",
							"Lead_Conversion_Time":   "Lead_Conversion_Time",
							"Lead_Source":            "Lead_Source",
							"Locked__s":              "Locked__s",
							"Modified_By":            "Modified_By",
							"Modified_Time":          "Modified_Time",
							"Next_Step":              "Next_Step",
							"Overall_Sales_Duration": "Overall_Sales_Duration",
							"Owner":                  "Owner",
							"Probability":            "Probability",
							"Reason_For_Loss__s":     "Reason_For_Loss__s",
							"Record_Image":           "Record_Image",
							"Record_Status__s":       "Record_Status__s",
							"Sales_Cycle_Duration":   "Sales_Cycle_Duration",
							"Stage":                  "Stage",
							"Type":                   "Type",
							"id":                     "id",
						},
					},
				},
				Errors: map[string]error{
					"Arsenal": common.NewHTTPStatusError(http.StatusBadRequest, fmt.Errorf("%w: %s", common.ErrCaller, string(arsenalResponse))),
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}

}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
