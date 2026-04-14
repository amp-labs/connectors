package chargeover

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

	okResponse := testutils.DataFromFile(t, "customers.json")
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
			Input: []string{"customers"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, okResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"superuser_id": {
								DisplayName:  "superuser_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"external_key": {
								DisplayName:  "external_key",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"token": {
								DisplayName:  "token",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"company": {
								DisplayName:  "company",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"terms_id": {
								DisplayName:  "terms_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"class_id": {
								DisplayName:  "class_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"custom_1": {
								DisplayName:  "custom_1",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"admin_id": {
								DisplayName:  "admin_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"campaign_id": {
								DisplayName:  "campaign_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"currency_id": {
								DisplayName:  "currency_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"language_id": {
								DisplayName:  "language_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"brand_id": {
								DisplayName:  "brand_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"default_paymethod": {
								DisplayName:  "default_paymethod",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"default_creditcard_id": {
								DisplayName:  "default_creditcard_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"default_ach_id": {
								DisplayName:  "default_ach_id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"tax_ident": {
								DisplayName:  "tax_ident",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"no_taxes": {
								DisplayName:  "no_taxes",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "boolean",
								Values:       nil,
							},
							"no_dunning": {
								DisplayName:  "no_dunning",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "boolean",
								Values:       nil,
							},
							"no_latefees": {
								DisplayName:  "no_latefees",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "boolean",
								Values:       nil,
							},
							"no_procfees": {
								DisplayName:  "no_procfees",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"write_datetime": {
								DisplayName:  "write_datetime",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"write_ipaddr": {
								DisplayName:  "write_ipaddr",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"mod_datetime": {
								DisplayName:  "mod_datetime",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"mod_ipaddr": {
								DisplayName:  "mod_ipaddr",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"customer_status_state": {
								DisplayName:  "customer_status_state",
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
		Workspace:           "test",
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
