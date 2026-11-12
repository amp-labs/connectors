package marketo

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	channelsResponse := testutils.DataFromFile(t, "channels.json")
	describeCompaniesResponse := testutils.DataFromFile(t, "describe-companies.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:   "Metadata from static file",
			Input:  []string{"activities"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities": {
						DisplayName: "activities",
						FieldsMap: map[string]string{
							"activityDate":            "activityDate",
							"activityTypeId":          "activityTypeId",
							"attributes":              "attributes",
							"campaignId":              "campaignId",
							"id":                      "id",
							"leadId":                  "leadId",
							"marketoGUID":             "marketoGUID",
							"primaryAttributeValue":   "primaryAttributeValue",
							"primaryAttributeValueId": "primaryAttributeValueId",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe supported object",
			Input: []string{"channels"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/rest/asset/v1/channels.json"),
					Then: mockserver.Response(http.StatusOK, channelsResponse),
				}},
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"channels": {
						DisplayName: "channels",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeFloat,
							},
							"name": {
								DisplayName: "name",
								ValueType:   common.ValueTypeString,
							},
							"progressionStatuses": {
								DisplayName: "progressionStatuses",
								ValueType:   common.ValueTypeOther,
							},
							"applicableProgramType": {
								DisplayName: "applicableProgramType",
								ValueType:   common.ValueTypeString,
							},
							"createdAt": {
								DisplayName: "createdAt",
								ValueType:   common.ValueTypeString,
							},
							"updatedAt": {
								DisplayName: "updatedAt",
								ValueType:   common.ValueTypeString,
							},
						},
						FieldsMap: map[string]string{
							"applicableProgramType": "applicableProgramType",
							"createdAt":             "createdAt",
							"id":                    "id",
							"name":                  "name",
							"progressionStatuses":   "progressionStatuses",
							"updatedAt":             "updatedAt",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe object having describe endpoint",
			Input: []string{"companies"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/rest/v1/companies/describe.json"),
					Then: mockserver.Response(http.StatusOK, describeCompaniesResponse),
				}},
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"companies": {
						DisplayName: "companies",
						Fields: map[string]common.FieldMetadata{
							"createdAt": {
								DisplayName:  "Created At",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "datetime",
								ReadOnly:     new(true),
							},
							"externalCompanyId": {
								DisplayName:  "External Company Id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								ReadOnly:     new(true),
							},
							"id": {
								DisplayName:  "Id",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								ReadOnly:     new(true),
							},
							"annualRevenue": {
								DisplayName:  "Annual Revenue",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "currency",
								ReadOnly:     new(false),
							},
							"billingCity": {
								DisplayName:  "Billing City",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								ReadOnly:     new(false),
							},
						},
						FieldsMap: map[string]string{
							"createdAt":         "Created At",
							"externalCompanyId": "External Company Id",
							"id":                "Id",
							"annualRevenue":     "Annual Revenue",
							"billingCity":       "Billing City",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithWorkspace("test-workspace"),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
