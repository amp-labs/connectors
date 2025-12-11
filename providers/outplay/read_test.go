package outplay

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint:funlen
func TestRead(t *testing.T) {
	t.Parallel()

	prospectAccountResponse := testutils.DataFromFile(t, "prospectaccount-read.json")
	callAnalysisResponse := testutils.DataFromFile(t, "callanalysis-read.json")

	tests := []testroutines.Read{
		{
			Name:         "Read objects must be implemented",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successful read for prospectaccount",
			Input: common.ReadParams{
				ObjectName: "prospectaccount",
				Fields:     connectors.Fields("accountid", "name", "description"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/api/v1/prospectaccount/search"),
					},
					Then: mockserver.Response(http.StatusOK, prospectAccountResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"accountid":   float64(16448),
						"name":        "padikal",
						"description": "HCL Technologies Limited is an Indian multinational information technology service and consulting company headquartered in Noida, Uttar Pradesh.", // nolint: lll
					},
					Raw: map[string]any{
						"stage": map[string]any{
							"id":   float64(0),
							"name": "",
						},
						"owner": map[string]any{
							"id":    float64(263),
							"name":  "Suresh Avatar",
							"email": "outplaytest@martinz.co.in",
						},
						"accountid":     float64(16448),
						"name":          "padikal",
						"externalid":    "",
						"description":   "HCL Technologies Limited is an Indian multinational information technology service and consulting company headquartered in Noida, Uttar Pradesh.", // nolint: lll
						"employeecount": float64(0),
						"industrytype":  "springs3",
						"linkedin":      "",
						"twitter":       "http://twitter.com/suresh",
						"foundedyear":   "1976",
						"city":          "Noida",
						"website":       "http://HCL.com",
						"fields": map[string]any{
							"acc_txtpicklist": "neww",
							"acc_nopicklist":  "",
							"acc_date":        "2022-08-11",
							"acc_textfield":   "",
							"acc_numfield":    "",
							"rating":          "",
							"farooqaccount":   "",
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read for callanalysis",
			Input: common.ReadParams{
				ObjectName: "callanalysis",
				Fields:     connectors.Fields("callmetadataid", "title", "callduration"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/v1/callanalysis/list"),
					},
					Then: mockserver.Response(http.StatusOK, callAnalysisResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"callmetadataid": float64(2367),
						"title":          "Outplay Training",
						"callduration":   float64(18),
					},
					Raw: map[string]any{
						"callmetadataid":    float64(2367),
						"title":             "Outplay Training",
						"callstarttime":     "02/06/2025 14:41:22",
						"callduration":      float64(18),
						"recordingfilepath": "https://ap1-app.somexyz.com/exportrecording/2367",
						"mimetype":          "audio/x-wav",
						"meetingtype":       "External",
						"callsource":        "GoogleMeet",
						"createddate":       "2025-02-06T14:42:38.682231",
						"attendees": []any{
							map[string]any{
								"Attendeeid": float64(4793372),
								"Name":       "",
								"Emailid":    "vaibhav.j@abcdef.co.in",
								"Type":       float64(2),
							},
						},
						"unknownattendees": []any{},
					},
				}},
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnectorForRead(tt.Server.URL)
			})
		})
	}
}

func constructTestConnectorForRead(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test",
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
