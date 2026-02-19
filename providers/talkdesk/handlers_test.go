package talkdesk

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

func constructTestConnector(serverURL string) (*Connector, error) {
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
func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	contacts := testutils.DataFromFile(t, "contacts.json")
	activities := testutils.DataFromFile(t, "activities.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successful read of contacts with chosen fields",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("phones", "name", "id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/contacts"),
				Then:  mockserver.Response(http.StatusOK, contacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"name": "Nad",
							"id":   "6752b05a23552310c15446cc",
							"phones": []any{
								map[string]any{
									"label":  "",
									"number": "+9779807463483",
								},
							},
						},
						Raw: map[string]any{
							"id":         "6752b05a23552310c15446cc",
							"name":       "Nad",
							"company":    "",
							"deleted_at": nil,
							"phones": []any{
								map[string]any{
									"label":  "",
									"number": "+9779807463483",
								},
							},
							"emails": []any{},
						},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read of activities with chosen fields",
			Input: common.ReadParams{ObjectName: "identity/activities", Fields: connectors.Fields("phone_valid", "phone_line_type", "id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/identity/activities"),
				Then:  mockserver.Response(http.StatusOK, activities),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":              float64(1889),
							"phone_valid":     true,
							"phone_line_type": "VoIP",
						},
						Raw: map[string]any{
							"id":                           float64(1889),
							"contact_id":                   "753",
							"interaction_id":               "c5572384a4c340c2b326965fdb13ca0e",
							"timestamp":                    "2022-08-23T10:36:17.803596Z",
							"phone_status":                 "SAFE",
							"phone_message":                "valid",
							"phone_success":                true,
							"phone_formatted":              "969-999-999",
							"phone_local_format":           "969-999-999",
							"phone_valid":                  true,
							"phone_risk_score":             float64(10),
							"phone_recent_abuse":           false,
							"phone_voip":                   true,
							"phone_prepaid":                false,
							"phone_risky":                  false,
							"phone_active":                 true,
							"phone_carrier":                "VODAFONE",
							"phone_line_type":              "VoIP",
							"phone_country":                "PT",
							"phone_region":                 "PT",
							"phone_dialing_code":           float64(351),
							"phone_request_id":             "requestId1",
							"phone_call_attestation_type":  "TN-Validation-Passed-C",
							"phone_call_attestation_valid": true,
							"voice_operation":              "VERIFY",
							"voice_risk_score":             float64(27.99999713897705),
							"voice_result":                 "SUCCESS",
							"overall_risk_score":           27.99999713897705,
							"is_deleted":                   false,
							"phone_number":                 "+351912345678",
							"voice_type":                   "PASSIVE",
						},
					},
				},
				Done:     false,
				NextPage: "https://api-docs.talkdesk.org/identity/activities?filter=voice_type%20eq%20%27PASSIVE%27%20and%20contact_id%20eq%20%27753%27%20and%20%20interaction_id%20eq%20%27c5572384a4c340c2b326965fdb13ca0e%27&page=3",
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
