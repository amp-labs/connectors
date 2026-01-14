package ringcentral

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	okResponse := testutils.DataFromFile(t, "meetings.json")
	unsupportedObjects := testutils.DataFromFile(t, "404.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},

		{
			Name:  "Server response must have at least one field",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedObjects),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrRetryable,
				},
			},
		},

		{
			Name:  "Successfully describe Meetings metadata",
			Input: []string{"meetings"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, okResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"meetings": {
						DisplayName: "Meetings",
						Fields: map[string]common.FieldMetadata{
							"bridgeId": {
								DisplayName:  "bridgeId",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"chatContentUrl": {
								DisplayName:  "chatContentUrl",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"displayName": {
								DisplayName:  "displayName",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"duration": {
								DisplayName:  "duration",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"hostInfo": {
								DisplayName:  "hostInfo",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"participants": {
								DisplayName:  "participants",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"shortId": {
								DisplayName:  "shortId",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"startTime": {
								DisplayName:  "startTime",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
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

	statesResponse := testutils.DataFromFile(t, "read-comm-handling-states.json")
	contactsResponse := testutils.DataFromFile(t, "contacts.json")

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
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("uri", "firstName", "lastName", "id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/restapi/v1.0/account/~/extension/~/address-book/contact"),
				Then:  mockserver.Response(http.StatusOK, contactsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"firstname": "Joseph",
						"id":        float64(490491053),
						"lastname":  "Karage",
						"uri":       "https://platform.ringcentral.com/restapi/v1.0/account/278988052/extension/278988052/address-book/contact/490491053",
					},
					Raw: map[string]any{
						"availability": "Alive",
						"company":      "Ampersand",
						"email":        "joseph.karage@withampersand.com",
						"firstName":    "Joseph",
						"id":           float64(490491053),
						"jobTitle":     "Contributor",
						"lastName":     "Karage",
						"mobilePhone":  "+255713507067",
						"uri":          "https://platform.ringcentral.com/restapi/v1.0/account/278988052/extension/278988052/address-book/contact/490491053",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read of states with chosen fields",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("enabled", "displayName", "conditions")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/restapi/v1.0/account/~/extension/~/address-book/contact"),
				Then:  mockserver.Response(http.StatusOK, statesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"conditions": []any{
							map[string]any{
								"schedule": map[string]any{
									"triggers": []any{
										map[string]any{
											"endTime":     "23:59:59",
											"startTime":   "00:00:00",
											"triggerType": "Daily",
										},
									},
								},
								"type": "Schedule",
							},
						},
						"displayname": "Forward all calls",
						"enabled":     false,
					},
					Raw: map[string]any{
						"conditions": []any{
							map[string]any{
								"schedule": map[string]any{
									"triggers": []any{
										map[string]any{
											"endTime":     "23:59:59",
											"startTime":   "00:00:00",
											"triggerType": "Daily",
										},
									},
								},
								"type": "Schedule",
							},
						},
						"displayName": "Forward all calls",
						"enabled":     false,
						"id":          "forward-all-calls",
					},
				}, {
					Fields: map[string]any{
						"conditions": []any{
							map[string]any{
								"schedule": map[string]any{
									"triggers": []any{
										map[string]any{
											"endTime":     "23:59:59",
											"startTime":   "00:00:00",
											"triggerType": "Daily",
										},
									},
								},
								"type": "Schedule",
							},
						},
						"displayname": "Do not disturb",
						"enabled":     false,
					},
					Raw: map[string]any{
						"conditions": []any{
							map[string]any{
								"schedule": map[string]any{
									"triggers": []any{
										map[string]any{
											"endTime":     "23:59:59",
											"startTime":   "00:00:00",
											"triggerType": "Daily",
										},
									},
								},
								"type": "Schedule",
							},
						},
						"displayName": "Do not disturb",
						"enabled":     false,
						"id":          "dnd",
					},
				}},
				NextPage: "",
				Done:     true,
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
