package batch

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/datautils"
)

// BatchWrite executes a Salesforce composite create or update request,
// depending on the parameters provided. It validates input, builds the proper
// payload, sends the API request, and parses the results into a BatchWriteResult structure.
func (a *Adapter) BatchWrite(ctx context.Context, params *common.BatchWriteParam) (*common.BatchWriteResult, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := a.buildBatchWriteURL(params)
	if err != nil {
		return nil, err
	}

	payload, err := buildBatchWritePayload(params)
	if err != nil {
		return nil, err
	}

	// Choose REST method.
	write := a.Client.Post
	if params.IsUpdate() {
		write = a.Client.Patch
	}

	// Make an API call.
	rsp, err := write(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	// TODO response for the UPDATE endpoint has different schema.

	// Parse and process response.
	response, err := common.UnmarshalJSON[Response](rsp)
	if err != nil {
		return nil, err
	}

	if response == nil {
		status := common.BatchStatusSuccess
		errors := make([]any, 0)

		if rsp.Code == http.StatusBadRequest {
			// A 400 Bad Request is allowed by implementation, but we always expect a response body.
			// Since there is no data, and non-2xx response we cannot determine per-record results,
			// so the batch is treated as failed.
			status = common.BatchStatusFailure
			errors = append(errors, common.ErrEmptyJSONHTTPResponse)
		}

		return &common.BatchWriteResult{
			Status:  status,
			Errors:  errors,
			Results: nil,
		}, nil
	}

	// Map indexed by unique reference ids. Created once for the lookup.
	items := response.GetItemsMap()

	return common.ParseBatchWrite(
		payload.Records,
		func(index int, payloadItem PayloadItem) *Item {
			return items[payloadItem.Extension.Attributes.ReferenceID]
		},
		constructWriteResult,
	)
}

func constructWriteResult(payloadItem PayloadItem, respItem *Item) (*common.WriteResult, error) {
	if respItem == nil {
		// Salesforce didn't return matching response for the record.
		// This only means that some other records have failed and no records were processed.
		// However, this record was valid.
		return &common.WriteResult{
			Success:  false, // not processed
			RecordId: "",
			Errors: []any{
				common.ErrBatchUnprocessedRecord,
				fmt.Sprintf("record's referenceId is %v", payloadItem.Extension.Attributes.ReferenceID),
			},
			Data: nil,
		}, nil
	}

	return respItem.ToWriteResult()
}

func (a *Adapter) buildBatchWriteURL(params *common.BatchWriteParam) (*urlbuilder.URL, error) {
	if params.IsCreate() {
		return a.getCreateURL(params.ObjectName)
	}

	if params.IsUpdate() {
		return a.getUpdateURL()
	}

	return nil, common.ErrUnsupportedBatchWriteType
}

func buildBatchWritePayload(params *common.BatchWriteParam) (*Payload, error) {
	records, err := params.GetRecords()
	if err != nil {
		return nil, err
	}

	items := make([]PayloadItem, len(records))
	for index, record := range records {
		items[index] = PayloadItem{
			Record: record,
			Extension: RecordExtension{
				Attributes: RecordAttributes{
					Type:        params.ObjectName.String(),
					ReferenceID: fmt.Sprintf("ref%d", index),
				},
			},
		}
	}

	return &Payload{Records: items}, nil
}

// Payload represents the composite API request body.
// Each record is wrapped in a PayloadItem that carries additional metadata.
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/dome_composite_sobject_tree_flat.htm
type Payload struct {
	Records []PayloadItem `json:"records"`
}

// PayloadItem represents a single item in the composite API payload.
// It wraps a core Record with Salesforce-specific attributes required
// for batch or composite write operations. Fields from RecordExtension
// are merged alongside the record's own properties in the final payload.
type PayloadItem = codec.DecoratedRecord[RecordExtension]

type RecordExtension struct {
	Attributes RecordAttributes `json:"attributes"`
}

type RecordAttributes struct {
	Type        string `json:"type"`
	ReferenceID string `json:"referenceId"`
}

// Response is structure returned by API either for "200 OK" or "400 Bad Request".
type Response struct {
	HasErrors bool   `json:"hasErrors"`
	Results   []Item `json:"results"`
}

type Item struct {
	ReferenceId string      `json:"referenceId"`
	ID          string      `json:"id"`
	Errors      []ItemError `json:"errors"`
}

type ItemError struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message"`
	Fields     []any  `json:"fields"`
}

func (r Response) GetItemsMap() map[string]*Item {
	mapping := make(map[string]*Item)

	for _, item := range r.Results {
		mapping[item.ReferenceId] = &item
	}

	return mapping
}

func (i Item) ToWriteResult() (*common.WriteResult, error) {
	data, err := common.RecordDataToMap(i)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  len(i.Errors) == 0,
		RecordId: i.ID,
		Errors:   datautils.ToAnySlice(i.Errors),
		Data:     data,
	}, nil
}
