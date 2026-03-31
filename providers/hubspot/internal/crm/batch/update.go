package batch

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/associations"
)

// buildBatchUpdatePayload constructs a batch update Payload.
// For all records that have associations defined, it also builds
// a BatchCreateParams instance that can be used for follow‑up association creation.
func buildBatchUpdatePayload(params *common.BatchWriteParam) (*Payload, *associations.BatchCreateParams, error) {
	payloadItems := make([]PayloadItem, len(params.Batch))
	batchCreateParams := associations.NewBatchCreateParams(params.ObjectName)

	for index, batchItem := range params.Batch {
		record, err := batchItem.GetRecord()
		if err != nil {
			return nil, nil, err
		}

		item, err := NewPayloadItem(record)
		if err != nil {
			return nil, nil, err
		}

		associationsList, err := associations.ParseInput(batchItem.Associations)
		if err != nil {
			return nil, nil, err
		}

		batchCreateParams.WithAssociation(item.ID, associationsList)

		payloadItems[index] = *item
	}

	return &Payload{Items: payloadItems}, batchCreateParams, nil
}
