package braintree

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

func TestListObjectMetadata(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	customersMetadataResponse := testutils.DataFromFile(t, "metadata-customers.json")
	metadataErrorResponse := testutils.DataFromFile(t, "metadata-error.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Error when introspection returns empty fields",
			Input: []string{"invalidObject"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/graphql"),
					mockcond.Header(http.Header{
						"Braintree-Version": []string{"2019-01-01"},
					}),
				},
				Then: mockserver.Response(http.StatusOK, metadataErrorResponse),
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"invalidObject": common.ErrMissingExpectedValues,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe customers object with metadata via GraphQL introspection",
			Input: []string{"customers"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/graphql"),
					mockcond.Header(http.Header{
						"Braintree-Version": []string{"2019-01-01"},
					}),
				},
				Then: mockserver.Response(http.StatusOK, customersMetadataResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "ID",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "String",
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "string",
								ProviderType: "String",
							},
							"lastName": {
								DisplayName:  "lastName",
								ValueType:    "string",
								ProviderType: "String",
							},
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
