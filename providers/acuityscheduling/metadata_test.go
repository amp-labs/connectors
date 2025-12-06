package acuityscheduling

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

	clientsResponse := testutils.DataFromFile(t, "clients-read.json")
	blocksResponse := testutils.DataFromFile(t, "blocks-read.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"clients", "blocks"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v1/clients"),
					Then: mockserver.Response(http.StatusOK, clientsResponse),
				}, {
					If:   mockcond.Path("/api/v1/blocks"),
					Then: mockserver.Response(http.StatusOK, blocksResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"clients": {
						DisplayName: "Clients",
						Fields: map[string]common.FieldMetadata{
							"firstName": {
								DisplayName:  "FirstName",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"lastName": {
								DisplayName:  "LastName",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"phone": {
								DisplayName:  "Phone",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"notes": {
								DisplayName:  "Notes",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"email":     "Email",
							"firstName": "FirstName",
							"lastName":  "LastName",
							"notes":     "Notes",
							"phone":     "Phone",
						},
					},
					"blocks": {
						DisplayName: "Blocks",
						Fields: map[string]common.FieldMetadata{
							"description": {
								DisplayName:  "Description",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"until": {
								DisplayName:  "Until",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"recurring": {
								DisplayName:  "Recurring",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"notes": {
								DisplayName:  "Notes",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"end": {
								DisplayName:  "End",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"start": {
								DisplayName:  "Start",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"calendarID": {
								DisplayName:  "CalendarID",
								ValueType:    "float",
								ProviderType: "float",
								Values:       nil,
							},
							"id": {
								DisplayName:  "Id",
								ValueType:    "float",
								ProviderType: "float",
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"calendarID":  "CalendarID",
							"description": "Description",
							"end":         "End",
							"id":          "Id",
							"notes":       "Notes",
							"recurring":   "Recurring",
							"start":       "Start",
							"until":       "Until",
						},
					},
				},
				Errors: map[string]error{},
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
