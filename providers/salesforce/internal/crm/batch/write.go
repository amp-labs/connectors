package batch

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
)

// nolint:lll
// BatchWrite executes a Salesforce composite create or update request.
// It validates the input, builds the appropriate payload, sends the API call,
// and parses the response into a BatchWriteResult.
//
// The payload formats for Create and Update endpoints are nearly identical.
// The only notable difference—unused in this implementation—is the optional
// "allOrNone" flag supported by the Update API (default is false).
// See: https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_update.htm
//
// Response schemas differ slightly: Create responses wrap records in an
// enclosing object, while Update responses return a list at the root level.
// Each item also varies subtly in shape, represented by distinct Go structs,
// though their nested error formats are identical.
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

	write := a.Client.Post
	if params.IsUpdate() {
		write = a.Client.Patch
	}

	rsp, err := write(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	// Parse and process response.
	response, err := common.UnmarshalJSON[Response](rsp)
	if err != nil {
		return nil, err
	}

	if response == nil || len(*response) == 0 {
		return a.handleEmptyResponse(rsp)
	}

	// nolint:lll
	return common.ParseBatchWrite(
		payload.Records,
		responseMatcher(response),
		resultBuilder,
		nil,
	)
}

// nolint:lll
// responseMatcher matches records with response items.
//
// The index is used to correlate payloads and responses.
// However, we still guard against out-of-range access to ensure robustness
// if the response length is shorter than expected.
//
// From the Salesforce docs:
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_create.htm
//
//	"Objects are created in the order they’re listed.
//	The SaveResult objects are returned in the order in which the create requests were specified."
//
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_update.htm
//
//	"Objects are updated in the order they’re listed.
//	 The SaveResult objects are returned in the same order."
func responseMatcher(response *Response) common.BatchWriteResponseMatcher[PayloadItem, Item] {
	return func(index int, payloadItem PayloadItem) *Item {
		list := *response
		if index < 0 || index >= len(list) {
			return nil
		}

		return &list[index]
	}
}

func resultBuilder(_ PayloadItem, respItem *Item) (*common.WriteResult, error) {
	if respItem == nil {
		return createUnprocessableItem(), nil
	}

	return respItem.ToWriteResult()
}

func (a *Adapter) handleEmptyResponse(rsp *common.JSONHTTPResponse) (*common.BatchWriteResult, error) {
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

func (a *Adapter) buildBatchWriteURL(params *common.BatchWriteParam) (*urlbuilder.URL, error) {
	if params.IsCreate() {
		return a.getCreateURL()
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
					Type: params.ObjectName.String(),
				},
			},
		}
	}

	return &Payload{
		Records:   items,
		AllOrNone: goutils.Pointer(true),
	}, nil
}

// Payload represents the composite API request body.
// Each record is wrapped in a PayloadItem that carries additional metadata.
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/dome_composite_sobject_tree_flat.htm
type Payload struct {
	Records []PayloadItem `json:"records"`

	// AllOrNone is accepted by Create and Update endpoint for output with partial success.
	AllOrNone *bool `json:"allOrNone"`
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
	Type string `json:"type"`
}

// Response is structure returned by API either for "200 OK" or "400 Bad Request".
type Response []Item

type Item struct {
	Success bool        `json:"success"`
	ID      string      `json:"id,omitempty"`
	Errors  []ItemError `json:"errors"`
}

type ItemError struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message"`
	Fields     []any  `json:"fields"`
}

func (i Item) ToWriteResult() (*common.WriteResult, error) {
	success := len(i.Errors) == 0

	if success {

		return &common.WriteResult{
			Success:  true,
			RecordId: i.ID,
			Errors:   nil,
			Data:     nil,
		}, nil
	}

	return &common.WriteResult{
		Success:  false,
		RecordId: i.ID,
		Errors:   datautils.ToAnySlice(i.Errors),
		Data:     nil,
	}, nil
}

func createUnprocessableItem() *common.WriteResult {
	// Salesforce didn't return matching response for the record.
	// This only means that some other records have failed and no records were processed.
	// However, this record was valid.
	return &common.WriteResult{
		Success:  false, // not processed
		RecordId: "",
		Errors: []any{
			common.ErrBatchUnprocessedRecord,
		},
		Data: nil,
	}
}
