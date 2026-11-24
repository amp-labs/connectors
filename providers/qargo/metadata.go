package qargo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/qargo/internal/metadata"
)

// listObjectMetadata serves the objects Read lists from the catalog generated from the OpenAPI spec.
// The payment status write is not a collection, so the generator has no list response to describe it;
// its one field is the documented request body, PaymentStatusUpdate.
// https://api-docs.qargo.com/docs/api-accounting ("Report Payment Status For Sales Invoice")
func (c *Connector) listObjectMetadata(
	_ context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := metadata.Catalog.SelectResult(objectNames)
	if err != nil {
		return nil, err
	}

	for _, objectName := range objectNames {
		if objectName == objectNameSalesInvoicePayment {
			delete(result.Errors, objectName)
			result.Result[objectName] = *salesInvoicePaymentMetadata()
		}
	}

	return result, nil
}

func salesInvoicePaymentMetadata() *common.ObjectMetadata {
	return common.NewObjectMetadata("Sales Invoice Payment", common.FieldsMetadata{
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
	})
}
