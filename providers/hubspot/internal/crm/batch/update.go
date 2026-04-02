package batch

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/associations"
)

func (a *Adapter) batchUpdate(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error) {
	url, err := a.getUpdateURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	payload, batchCreateParams, err := buildBatchUpdatePayload(params)
	if err != nil {
		return nil, err
	}

	// Make an API call.
	rsp, err := a.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	if httpkit.Status4xx(rsp.Code) {
		// 4xx responses (e.g., 400 or 409) represent valid request outcomes
		// that include structured issue details, not fatal API failures.
		// Critical errors (5xx and the rest of 4xx) are handled by the HTTP client and returned as Go errors.
		return parseBulkIssue(payload, rsp)
	}

	// Associations can be created as part of batch create but not as part of batch update.
	// Therefore, batch update requires a dedicated follow‑up association request.
	associationsResult, err := a.associationsStrategy.BatchCreate(ctx, batchCreateParams)
	if err != nil {
		return nil, err
	}

	bulkResponse, err := parseBulkResponse(params, payload, rsp)
	if err != nil {
		return nil, err
	}

	// For batch update, association creation is handled in a separate call.
	// If that call failed, collect the errors as warnings rather than failing the whole operation.
	if associationsResult != nil && !associationsResult.Success {
		bulkResponse.Errors = append(bulkResponse.Errors, associationsResult.Errors)
	}

	return bulkResponse, nil
}

// buildBatchUpdatePayload constructs a batch update Payload.
// For all records that have associations defined, it also builds
// a BatchCreateParams instance that can be used for follow‑up association creation.
// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/deals/batch/create-deals
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
