package aws

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Metadata for objects across AWS Services",
			Input:      []string{"Users", "Applications", "Helicopters"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"IdentityStoreId":   "Identity Store Id",
							"UserId":            "User Id",
							"Addresses":         "Addresses",
							"DisplayName":       "Display Name",
							"Emails":            "Emails",
							"ExternalIds":       "External Ids",
							"Locale":            "Locale",
							"Name":              "Name",
							"NickName":          "NickName",
							"PhoneNumbers":      "Phone Numbers",
							"PreferredLanguage": "Preferred Language",
							"ProfileUrl":        "Profile URL",
							"Timezone":          "Timezone",
							"Title":             "Title",
							"UserName":          "User Name",
							"UserType":          "User Type",
						},
					},
					"Applications": {
						DisplayName: "Applications",
						FieldsMap: map[string]string{
							"ApplicationAccount":     "Application Account",
							"ApplicationArn":         "Application Arn",
							"ApplicationProviderArn": "Application Provider Arn",
							"CreatedDate":            "Created Date",
							"Description":            "Description",
							"InstanceArn":            "Instance Arn",
							"Name":                   "Name",
							"PortalOptions":          "Portal Options",
							"Status":                 "Status",
						},
					},
				},
				Errors: map[string]error{
					"Helicopters": common.ErrObjectNotSupported,
				},
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
	connector, err := NewConnector(common.ConnectorParams{
		Module:              providers.ModuleAWSIdentityCenter,
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"region":          "test-region",
			"identityStoreId": "test-identity-store-id",
			"instanceARN":     "test-instance-arn",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.getModuleURL(), serverURL))

	return connector, nil
}

func TestURLDomain(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(common.ConnectorParams{
		Module:              providers.ModuleAWSIdentityCenter,
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"region":          "test-region",
			"identityStoreId": "test-identity-store-id",
			"instanceARN":     "test-instance-arn",
		},
	})
	if err != nil {
		t.Fatalf("failed to build AWS connector: %v", err)
	}

	moduleURL := connector.getModuleURL()

	testutils.CheckOutput(t, "AWS domain substitution",
		"https://<<SERVICE_DOMAIN>>.test-region.amazonaws.com", moduleURL)
}
