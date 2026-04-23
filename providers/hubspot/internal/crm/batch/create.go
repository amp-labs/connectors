package batch

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (a *Adapter) batchCreate(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error) {
	url, err := a.getCreateURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	payload, err := buildBatchCreatePayload(params)
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

	return parseBulkResponse(params, payload, rsp)
}

// https://developers.hubspot.com/docs/api-reference/latest/crm/objects/deals/batch/create-deals
func buildBatchCreatePayload(params *common.BatchWriteParam) (*Payload, error) {
	payloadItems := make([]PayloadItem, len(params.Batch))

	// For creates, include objectWriteTraceId only when allOrNone is false (default).
	// This enables partial success - HubSpot returns trace IDs in error responses for per-record matching.
	// When allOrNone is true, omit objectWriteTraceId so HubSpot fails the entire batch if any record fails.
	includeTraceId := !params.GetAllOrNone()

	for index, batchItem := range params.Batch {
		record, err := batchItem.GetRecord()
		if err != nil {
			return nil, err
		}

		item, err := NewPayloadItem(record)
		if err != nil {
			return nil, err
		}

		// Prepare associations for the payload.
		item.Associations = batchItem.Associations

		if includeTraceId {
			item.ObjectWriteTraceId = formatTraceId(index)
		}

		payloadItems[index] = *item
	}

	return &Payload{Items: payloadItems}, nil
}

// Payload represents the HubSpot batch request body.
type Payload struct {
	Items []PayloadItem `json:"inputs"`
}

// PayloadItem represents a single item in the API payload.
// Hubspot's payload is identical to what client supplies to the connector.
// This is an alias.
type PayloadItem struct {
	ID                 string        `json:"id,omitempty"`
	Properties         common.Record `json:"properties"`
	Associations       any           `json:"associations,omitempty"`
	ObjectWriteTraceId string        `json:"objectWriteTraceId,omitempty"` //nolint:tagliatelle
}

func NewPayloadItem(record common.Record) (*PayloadItem, error) {
	node, err := jsonquery.Convertor.NodeFromMap(record)
	if err != nil {
		return nil, err
	}

	identifier, err := jsonquery.New(node).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	properties, err := datautils.FromMap(record).DeepCopy()
	if err != nil {
		return nil, err
	}

	// Hubspot will complain about unexpected fields, must clean up
	delete(properties, "id")

	return &PayloadItem{
		ID:         identifier,
		Properties: common.Record(properties),
	}, nil
}

// formatTraceId creates a trace ID string from an index.
// Used for objectWriteTraceId in HubSpot create operations.
func formatTraceId(index int) string {
	return strconv.Itoa(index)
}
