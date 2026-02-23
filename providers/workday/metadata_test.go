package workday

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "Unknown object requested",
			Input:        []string{"nonexistent_object"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent_object": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:         "Successfully describe workers object with metadata",
			Input:        []string{"workers"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workers": {
						DisplayName: "Workers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"descriptor": {
								DisplayName:  "descriptor",
								ValueType:    "string",
								ProviderType: "string",
							},
							"primaryWorkEmail": {
								DisplayName:  "primaryWorkEmail",
								ValueType:    "string",
								ProviderType: "string",
							},
							"isManager": {
								DisplayName:  "isManager",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe organizations object with metadata",
			Input:        []string{"organizations"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"organizations": {
						DisplayName: "Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"descriptor": {
								DisplayName:  "descriptor",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe supervisory organizations object with metadata",
			Input:        []string{"supervisoryOrganizations"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"supervisoryOrganizations": {
						DisplayName: "Supervisory Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe currencies object with metadata",
			Input:        []string{"currencies"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"currencies": {
						DisplayName: "Currencies",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"precision": {
								DisplayName:  "precision",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe multiple objects with metadata",
			Input:        []string{"workers", "organizations", "auditLogs", "jobChangeReasons"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workers": {
						DisplayName: "Workers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"organizations": {
						DisplayName: "Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"auditLogs": {
						DisplayName: "Audit Logs",
						Fields: map[string]common.FieldMetadata{
							"activityAction": {
								DisplayName:  "activityAction",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"jobChangeReasons": {
						DisplayName: "Job Change Reasons",
						Fields: map[string]common.FieldMetadata{
							"isForEmployee": {
								DisplayName:  "isForEmployee",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
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
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test",
		Metadata: map[string]string{
			"tenantName": "testTenant",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
