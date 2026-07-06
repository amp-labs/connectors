package dynamicsbusiness

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCustomers := testutils.DataFromFile(t, "metadata/customers.json")

	tests := []testconn.TestCaseListObjectMetadata{
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
			Comparator: testconn.ComparatorSubsetMetadata,
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
						common.ErrRetryable,
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

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
		Workspace:           "test-workspace",
		Metadata: map[string]string{
			"companyId":       "test-company-id",
			"environmentName": "test-environment",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseURL(server.URL)

	return connector, nil
}
