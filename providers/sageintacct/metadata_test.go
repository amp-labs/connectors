package sageintacct

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
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
							// Top-level fields
							"$['accountEmail']": {
								DisplayName:  "$['Accountemail']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['adminPrivileges']": {
								DisplayName:  "$['Adminprivileges']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
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
							// Nested fields from refs.contact
							"$['contact']['id']": {
								DisplayName:  "$['Contact']['Id']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['contact']['key']": {
								DisplayName:  "$['Contact']['Key']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['contact']['firstName']": {
								DisplayName:  "$['Contact']['Firstname']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['contact']['lastName']": {
								DisplayName:  "$['Contact']['Lastname']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['contact']['email1']": {
								DisplayName:  "$['Contact']['Email1']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							// Nested fields from refs.contact.groups.mailingAddress
							"$['$['contact']['mailingAddress']']['addressLine1']": {
								DisplayName:  "$['$['Contact']['Mailingaddress']']['Addressline1']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['$['contact']['mailingAddress']']['city']": {
								DisplayName:  "$['$['Contact']['Mailingaddress']']['City']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['$['contact']['mailingAddress']']['state']": {
								DisplayName:  "$['$['Contact']['Mailingaddress']']['State']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['$['contact']['mailingAddress']']['country']": {
								DisplayName:  "$['$['Contact']['Mailingaddress']']['Country']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							// Nested fields from refs.entity
							"$['entity']['id']": {
								DisplayName:  "$['Entity']['Id']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['entity']['key']": {
								DisplayName:  "$['Entity']['Key']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(false),
								Values:       []common.FieldValue{},
							},
							"$['entity']['name']": {
								DisplayName:  "$['Entity']['Name']",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(true),
								Values:       []common.FieldValue{},
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
