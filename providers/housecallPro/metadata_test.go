package housecallpro

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successfully describe material categories, job types, and invoices",
			Input:      []string{"price_book/material_categories", "job_fields/job_types", "invoices"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"price_book/material_categories": {
						DisplayName: "Material Categories",
						Fields: map[string]common.FieldMetadata{
							"object": {
								DisplayName:  "object",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"image": {
								DisplayName:  "image",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"uuid": {
								DisplayName:  "uuid",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"parent_uuid": {
								DisplayName:  "parent_uuid",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
						},
					},
					"job_fields/job_types": {
						DisplayName: "Job Types",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
						},
					},
					"invoices": {
						DisplayName: "Invoices",
						Fields: map[string]common.FieldMetadata{
							"amount": {
								DisplayName:  "amount",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								Values:       nil,
							},
							"discounts": {
								DisplayName:  "discounts",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
								Values:       nil,
							},
							"display_due_concept": {
								DisplayName:  "display_due_concept",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"due_amount": {
								DisplayName:  "due_amount",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								Values:       nil,
							},
							"due_at": {
								DisplayName:  "due_at",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"due_concept": {
								DisplayName:  "due_concept",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"invoice_date": {
								DisplayName:  "invoice_date",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"invoice_number": {
								DisplayName:  "invoice_number",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"items": {
								DisplayName:  "items",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
								Values:       nil,
							},
							"job_id": {
								DisplayName:  "job_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"paid_at": {
								DisplayName:  "paid_at",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"payments": {
								DisplayName:  "payments",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
								Values:       nil,
							},
							"sent_at": {
								DisplayName:  "sent_at",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"service_date": {
								DisplayName:  "service_date",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"status": {
								DisplayName:  "status",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"subtotal": {
								DisplayName:  "subtotal",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								Values:       nil,
							},
							"taxes": {
								DisplayName:  "taxes",
								ValueType:    common.ValueTypeOther,
								ProviderType: "array",
								Values:       nil,
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
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

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
