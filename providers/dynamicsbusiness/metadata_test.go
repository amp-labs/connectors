package dynamicsbusiness

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCustomers := testutils.DataFromFile(t, "metadata/customers.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"customers", "mailboxes"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/v2.0/test-workspace/test-environment/api/v2.0/entityDefinitions"),
						mockcond.QueryParam("$filter", "entityName eq 'customer'"),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomers),
				}, {
					If: mockcond.And{
						mockcond.Path("/v2.0/test-workspace/test-environment/api/v2.0/entityDefinitions"),
						mockcond.QueryParam("$filter", "entityName eq 'mailbox'"),
					},
					Then: mockserver.ResponseString(http.StatusNotFound, "{}"),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"addressLine1": {DisplayName: "Address Line 1"},
							"city":         {DisplayName: "City"},
							"displayName":  {DisplayName: "Display Name"},
							"number":       {DisplayName: "No."},
						},
					},
				},
				Errors: map[string]error{
					"mailboxes": mockutils.ExpectedSubsetErrors{
						ErrMetadataNotFound,
					},
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
		AuthenticatedClient: http.DefaultClient,
		Workspace:           "test-workspace",
		Metadata: map[string]string{
			"companyId":       "test-company-id",
			"environmentName": "test-environment",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(serverURL)

	return connector, nil
}
