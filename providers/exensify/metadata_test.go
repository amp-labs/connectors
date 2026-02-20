package exensify

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

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	responsePolicies := testutils.DataFromFile(t, "read-policy.json")
	responseAuthError := testutils.DataFromFile(t, "error-auth.json")
	responseEmptyList := testutils.DataFromFile(t, "empty-policy-list.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "API error is stored per object in metadata errors",
			Input: []string{"policy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAuthError),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"policy": common.ErrRequestFailed,
				},
			},
		},
		{
			Name:  "Empty policy list returns missing expected values error",
			Input: []string{"policy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseEmptyList),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"policy": common.ErrMissingExpectedValues,
				},
			},
		},
		{
			Name:  "Successfully describe policy object metadata",
			Input: []string{"policy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responsePolicies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"policy": {
						DisplayName: "Policy",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"name": {
								DisplayName: "name",
								ValueType:   common.ValueTypeString,
							},
							"outputCurrency": {
								DisplayName: "outputCurrency",
								ValueType:   common.ValueTypeString,
							},
							"type": {
								DisplayName: "type",
								ValueType:   common.ValueTypeString,
							},
							"owner": {
								DisplayName: "owner",
								ValueType:   common.ValueTypeString,
							},
							"autoReporting": {
								DisplayName: "autoReporting",
								ValueType:   common.ValueTypeBoolean,
							},
							"requiresCategory": {
								DisplayName: "requiresCategory",
								ValueType:   common.ValueTypeBoolean,
							},
							"employees": {
								DisplayName: "employees",
								ValueType:   common.ValueTypeFloat,
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
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server.
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
