package sageintacct

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUserSchema := testutils.DataFromFile(t, "user-metadata.json")
	responseUnsupportedObject := testutils.DataFromFile(t, "unsupported-object-metadata.json")

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
				Always: mockserver.Response(http.StatusOK, responseUnsupportedObject),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},

		{
			Name:  "Successfully describe User metadata",
			Input: []string{"company-config/user"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUserSchema),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"company-config/user": {
						DisplayName: "Company-config/user",
						Fields: map[string]common.FieldMetadata{
							"accountEmail": {
								DisplayName:  "Accountemail",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       []common.FieldValue{},
							},
							"adminPrivileges": {
								DisplayName:  "Adminprivileges",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values: []common.FieldValue{
									{
										DisplayValue: "Off",
										Value:        "off",
									},
									{
										DisplayValue: "Limited",
										Value:        "limited",
									},
									{
										DisplayValue: "Full",
										Value:        "full",
									},
								},
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
