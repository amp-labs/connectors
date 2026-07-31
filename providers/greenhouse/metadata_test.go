package greenhouse

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:       "Unknown object requested",
			Input:      []string{"pipelines"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"pipelines": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe users object with metadata",
			Input:      []string{"users"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"site_admin": {
								DisplayName:  "site_admin",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"emails": {
								DisplayName:  "emails",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"applications", "scorecards"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"applications": {
						DisplayName: "Applications",
						Fields: map[string]common.FieldMetadata{
							"candidate_id": {
								DisplayName:  "candidate_id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{
									{Value: "rejected", DisplayValue: "rejected"},
									{Value: "hired", DisplayValue: "hired"},
									{Value: "converted", DisplayValue: "converted"},
									{Value: "in_process", DisplayValue: "in_process"},
								},
							},
							"prospect": {
								DisplayName:  "prospect",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"scorecards": {
						DisplayName: "Scorecards",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
				},
			},
		},
		{
			Name:       "Successfully describe candidates object",
			Input:      []string{"candidates"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"candidates": {
						DisplayName: "Candidates",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"private": {
								DisplayName:  "private",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"email_addresses": {
								DisplayName:  "email_addresses",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
				},
			},
		},
		{
			// Greenhouse spells the object name in lower case, while the display name
			// is an acronym which title casing cannot produce.
			Name:       "Acronym object retains its display name",
			Input:      []string{"eeoc"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"eeoc": {DisplayName: "EEOC"},
				},
			},
		},
		{
			// Select fields must expose the values they accept.
			// https://harvestdocs.greenhouse.io/reference/get_v3-jobs
			Name:       "Select field on jobs reports its possible values",
			Input:      []string{"jobs"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs": {
						DisplayName: "Jobs",
						Fields: map[string]common.FieldMetadata{
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{
									{Value: "open", DisplayValue: "open"},
									{Value: "draft", DisplayValue: "draft"},
									{Value: "closed", DisplayValue: "closed"},
								},
							},
							"confidential": {
								DisplayName:  "confidential",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
				},
			},
		},
		{
			// Search variants duplicate the plain list endpoints and are deliberately
			// excluded, see ignoreEndpoints in scripts/openapi/greenhouse/metadata.
			Name:       "Excluded search variant is not supported",
			Input:      []string{"candidates/search"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"candidates/search": common.ErrObjectNotSupported,
				},
			},
		},
		{
			// Custom field definitions describe metadata rather than record data.
			Name:       "Excluded custom field definitions are not supported",
			Input:      []string{"custom_fields"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"custom_fields": common.ErrObjectNotSupported,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
