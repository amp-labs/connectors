package heyreach

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Successfully describe multiple object with metadata",
			Input:      []string{"campaign", "list", "li_account"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaign": {
						DisplayName: "campaign",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"creationTime": {
								DisplayName:  "creationTime",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"name":         "name",
							"creationTime": "creationTime",
							"status":       "status",
						},
					},
					"list": {
						DisplayName: "list",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"creationTime": {
								DisplayName:  "creationTime",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"listType": {
								DisplayName:  "listType",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"name":         "name",
							"creationTime": "creationTime",
							"listType":     "listType",
						},
					},
					"li_account": {
						DisplayName: "li_account",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"emailAddress": {
								DisplayName:  "emailAddress",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"lastName": {
								DisplayName:  "lastName",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"emailAddress": "emailAddress",
							"firstName":    "firstName",
							"lastName":     "lastName",
						},
					},
				},
				Errors: nil,
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
	connector, err := NewConnector(common.Parameters{
		Module:              staticschema.RootModuleID,
		AuthenticatedClient: http.DefaultClient,
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(serverURL)

	return connector, nil
}
