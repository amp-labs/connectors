package procore

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

const (
	testCompanyID = "4283186"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	projectsResponse := testutils.DataFromFile(t, "projects.json")
	operationsResponse := testutils.DataFromFile(t, "operations.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Company-scoped v1.0 endpoint returns bare array",
			Input: []string{"projects"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects"),
					mockcond.QueryParam("per_page", "1"),
					mockcond.Header(http.Header{"Procore-Company-Id": []string{testCompanyID}}),
				},
				Then: mockserver.Response(http.StatusOK, projectsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"projects": {
						DisplayName: "Projects",
						Fields: map[string]common.FieldMetadata{
							"id":             {DisplayName: "id", ValueType: common.ValueTypeFloat, ProviderType: "float"},
							"name":           {DisplayName: "name", ValueType: common.ValueTypeString, ProviderType: "string"},
							"project_number": {DisplayName: "project_number", ValueType: common.ValueTypeString, ProviderType: "string"},
							"active":         {DisplayName: "active", ValueType: common.ValueTypeBoolean, ProviderType: "boolean"},
							"updated_at":     {DisplayName: "updated_at", ValueType: common.ValueTypeString, ProviderType: "string"},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "v2.0 endpoint wraps records under data key",
			Input: []string{"operations"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/v2.0/companies/" + testCompanyID + "/async_operations"),
					mockcond.QueryParam("per_page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, operationsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"operations": {
						DisplayName: "Operations",
						Fields: map[string]common.FieldMetadata{
							"id":             {DisplayName: "id", ValueType: common.ValueTypeString, ProviderType: "string"},
							"status":         {DisplayName: "status", ValueType: common.ValueTypeString, ProviderType: "string"},
							"operation_type": {DisplayName: "operation_type", ValueType: common.ValueTypeString, ProviderType: "string"},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
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
		Metadata:            map[string]string{metadataKeyCompany: testCompanyID},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
