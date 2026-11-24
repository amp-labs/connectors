package qargo

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object is reported per object",
			Input:      []string{"orders"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"orders": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
		{
			Name:       "Companies come from the generated catalog",
			Input:      []string{"company"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"company": {
						DisplayName: "Company",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"is_customer": {
								DisplayName:  "Is Customer",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
							"timestamp_updated": {
								DisplayName:  "Timestamp Updated",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "Status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "PROSPECT", DisplayValue: "PROSPECT"},
									{Value: "NOT_APPROVED", DisplayValue: "NOT_APPROVED"},
									{Value: "APPROVED", DisplayValue: "APPROVED"},
									{Value: "WARNING", DisplayValue: "WARNING"},
									{Value: "BLOCKED", DisplayValue: "BLOCKED"},
								},
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:       "Queue documents are described by their component schemas",
			Input:      []string{"sales-invoice", "sales-credit-note"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"sales-invoice": {
						DisplayName: "Sales Invoice",
						Fields: map[string]common.FieldMetadata{
							"due_date": {
								DisplayName:  "Due Date",
								ValueType:    common.ValueTypeDate,
								ProviderType: "string",
							},
							"line_items": {
								DisplayName:  "Line Items",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
							},
						},
					},
					"sales-credit-note": {
						DisplayName: "Sales Credit Note",
						Fields: map[string]common.FieldMetadata{
							"billing_entity": {
								DisplayName:  "Billing Entity",
								ValueType:    common.ValueTypeOther,
								ProviderType: "object",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:       "Payment status describes the documented request body",
			Input:      []string{"sales-invoice/payment"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"sales-invoice/payment": {
						DisplayName: "Sales Invoice Payment",
						Fields: map[string]common.FieldMetadata{
							"status": {
								DisplayName:  "Status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "PaymentStatus",
								IsRequired:   new(true),
								Values: []common.FieldValue{
									{Value: "UNPAID", DisplayValue: "Unpaid"},
									{Value: "PAID", DisplayValue: "Paid"},
								},
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
