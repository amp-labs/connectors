package square

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

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	customersResponse := testutils.DataFromFile(t, "customers.json")
	catalogResponse := testutils.DataFromFile(t, "catalog.json")
	merchantsResponse := testutils.DataFromFile(t, "merchants.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object is not supported",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": mockutils.ExpectedSubsetErrors{
						common.ErrObjectNotSupported,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe customers by sampling first record",
			Input: []string{"customers"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/customers"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, customersResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"created_at": {
								DisplayName: "created_at",
								ValueType:   common.ValueTypeString,
							},
							"version": {
								DisplayName: "version",
								ValueType:   common.ValueTypeFloat,
							},
							"preferences": {
								DisplayName: "preferences",
								ValueType:   common.ValueTypeOther,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe catalog by sampling objects array",
			Input: []string{"catalog"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/catalog/list"),
				Then:  mockserver.Response(http.StatusOK, catalogResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"catalog": {
						DisplayName: "Catalog",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"type": {
								DisplayName: "type",
								ValueType:   common.ValueTypeString,
							},
							"is_deleted": {
								DisplayName: "is_deleted",
								ValueType:   common.ValueTypeBoolean,
							},
							"item_data": {
								DisplayName: "item_data",
								ValueType:   common.ValueTypeOther,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe merchants from the singular merchant array key",
			Input: []string{"merchants"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/merchants"),
				Then:  mockserver.Response(http.StatusOK, merchantsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"merchants": {
						DisplayName: "Merchants",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   common.ValueTypeString,
							},
							"business_name": {
								DisplayName: "business_name",
								ValueType:   common.ValueTypeString,
							},
							"status": {
								DisplayName: "status",
								ValueType:   common.ValueTypeString,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when server responds with 500",
			Input: []string{"customers"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/customers"),
				Then:  mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"customers": mockutils.ExpectedSubsetErrors{
						common.ErrServer,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when records array is empty",
			Input: []string{"customers"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/customers"),
				Then:  mockserver.Response(http.StatusOK, []byte(`{"customers": []}`)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"customers": mockutils.ExpectedSubsetErrors{
						common.ErrMissingExpectedValues,
					},
				},
			},
			ExpectedErrs: nil,
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
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
