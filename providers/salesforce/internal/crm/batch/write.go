package batch

import (
	"context"
	"fmt"
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

	if params.IsUpdate() {
		return a.batchWriteUpdate(ctx, url, payload)
	}

	return a.batchWriteCreate(ctx, url, payload)
}

func (a *Adapter) batchWriteCreate(
	ctx context.Context, url *urlbuilder.URL, payload *Payload,
) (*common.BatchWriteResult, error) {
	rsp, err := a.Client.Post(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	// Parse and process response.
	response, err := common.UnmarshalJSON[ResponseCreate](rsp)
	if err != nil {
		// Check if this is an error response (e.g., allOrNone failure)
		if errorResult := a.handleErrorResponse(rsp, payload.Records); errorResult != nil {
			// Return both result and error so server returns 422
			return errorResult, fmt.Errorf("batch write failed: %d records failed", errorResult.FailureCount)
		}
		return nil, err
	}

	if response == nil {
		return a.handleEmptyResponse(rsp)
	}

	// Map indexed by unique reference ids. Created once for the lookup.
	items := response.GetItemsMap()

	result, err := common.ParseBatchWrite[PayloadItem, CreateItem](
		payload.Records,
		func(index int, payloadItem PayloadItem) *CreateItem {
			return items[payloadItem.Extension.Attributes.ReferenceID]
		},
		func(payloadItem PayloadItem, respItem *CreateItem) (*common.WriteResult, error) {
			if respItem == nil {
				return createUnprocessableItem(payloadItem), nil
			}

			return respItem.ToWriteResult()
		},
	)
	if err != nil {
		return nil, err
	}

	// For creates, allOrNone is always true by default (cannot be changed).
	// If there are failures, return 422
	if result.FailureCount > 0 {
		return result, fmt.Errorf("batch write failed: %d records failed", result.FailureCount)
	}

	return result, nil
}

func (a *Adapter) batchWriteUpdate(
	ctx context.Context, url *urlbuilder.URL, payload *Payload,
) (*common.BatchWriteResult, error) {
	rsp, err := a.Client.Patch(ctx, url.String(), payload)
	if err != nil {
		return nil, err
	}

	// Parse and process response.
	response, err := common.UnmarshalJSON[ResponseUpdate](rsp)
	if err != nil {
		// Check if this is an error response (e.g., allOrNone failure)
		if errorResult := a.handleErrorResponse(rsp, payload.Records); errorResult != nil {
			// Return both result and error so server returns 422
			return errorResult, fmt.Errorf("batch write failed: %d records failed", errorResult.FailureCount)
		}
		return nil, err
	}

	if response == nil {
		return a.handleEmptyResponse(rsp)
	}

	// nolint:lll
	result, err := common.ParseBatchWrite(
		payload.Records,
		func(index int, payloadItem PayloadItem) *UpdateItem {
			// In Salesforce composite update responses, each item corresponds
			// positionally to the submitted payload item. Even when a record fails,
			// its response entry is still present but may have an empty "id" field.
			//
			// The index is used to correlate payloads and responses. However, we still
			// guard against out-of-range access to ensure robustness if the response
			// length is shorter than expected.
			//
			// From the Salesforce docs:
			// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_update.htm
			//   "Objects are updated in the order they're listed.
			//    The SaveResult objects are returned in the same order."
			list := *response
			if index < 0 || index >= len(list) {
				return nil
			}

			return &list[index]
		},
		func(payloadItem PayloadItem, respItem *UpdateItem) (*common.WriteResult, error) {
			if respItem == nil {
				return createUnprocessableItem(payloadItem), nil
			}

			return respItem.ToWriteResult()
		},
	)
	if err != nil {
		return nil, err
	}

	// For updates, AllOrNone is set to true in the payload.
	// If there are failures, return 422
	if payload.AllOrNone != nil && *payload.AllOrNone && result.FailureCount > 0 {
		return result, fmt.Errorf("batch write failed: %d records failed", result.FailureCount)
	}

	return result, nil
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

// handleErrorResponse attempts to parse the response as a Salesforce error array.
// When allOrNone=true fails, Salesforce returns 400 with an error array instead of normal response.
// This function creates a BatchWriteResult with all records marked as failed.
// Returns nil if the response is not an error array.
func (a *Adapter) handleErrorResponse(rsp *common.JSONHTTPResponse, records []PayloadItem) *common.BatchWriteResult {
	// Try to unmarshal as error array using the common.UnmarshalJSON function
	sfErrors, err := common.UnmarshalJSON[[]SalesforceError](rsp)
	if err != nil || sfErrors == nil {
		// Not an error array, let the caller handle it
		return nil
	}

	if len(*sfErrors) == 0 {
		return nil
	}

	// Create failed WriteResult for each record
	results := make([]common.WriteResult, len(records))
	for i := range records {
		errors := make([]any, len(*sfErrors))
		for j, sfErr := range *sfErrors {
			errors[j] = ItemError{
				StatusCode: sfErr.ErrorCode,
				Message:    sfErr.Message,
				Fields:     []any{},
			}
		}

		results[i] = common.WriteResult{
			Success:  false,
			RecordId: "",
			Errors:   errors,
			Data:     nil,
		}
	}

	return &common.BatchWriteResult{
		Status:       common.BatchStatusFailure,
		Errors:       nil,
		Results:      results,
		SuccessCount: 0,
		FailureCount: len(records),
	}
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

	if params.IsUpdate() {
		return &Payload{
			Records:   items,
			AllOrNone: goutils.Pointer(true),
		}, nil
	}

	return &Payload{
		Records: items,
	}, nil
}

// Payload represents the composite API request body.
// Each record is wrapped in a PayloadItem that carries additional metadata.
// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/dome_composite_sobject_tree_flat.htm
type Payload struct {
	Records []PayloadItem `json:"records"`

	// AllOrNone is accepted by Update endpoint for output with partial success.
	// nolint:lll
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_composite_sobjects_collections_update.htm
	AllOrNone *bool `json:"allOrNone,omitempty"`
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

// ResponseCreate is structure returned by API either for "200 OK" or "400 Bad Request".
type ResponseCreate struct {
	HasErrors bool         `json:"hasErrors"`
	Results   []CreateItem `json:"results"`
}

// ResponseUpdate is structure retuned by update operation.
// This differs from the creation such that it is a list of objects at top JSON node level.
type ResponseUpdate []UpdateItem

type UpdateItem struct {
	Success bool        `json:"success"`
	ID      string      `json:"id,omitempty"`
	Errors  []ItemError `json:"errors"`
}

type CreateItem struct {
	ReferenceId string      `json:"referenceId"`
	ID          string      `json:"id"`
	Errors      []ItemError `json:"errors"`
}

type ItemError struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message"`
	Fields     []any  `json:"fields"`
}

// SalesforceError represents the error format returned by Salesforce API.
// This is used when Salesforce returns an error array instead of normal response.
type SalesforceError struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

func (r ResponseCreate) GetItemsMap() map[string]*CreateItem {
	mapping := make(map[string]*CreateItem)

	for _, item := range r.Results {
		mapping[item.ReferenceId] = &item
	}

	return mapping
}

func (i CreateItem) ToWriteResult() (*common.WriteResult, error) {
	success := len(i.Errors) == 0

	if success {
		data, err := common.RecordDataToMap(i)
		if err != nil {
			return nil, err
		}

		return &common.WriteResult{
			Success:  true,
			RecordId: i.ID,
			Errors:   nil,
			Data:     data,
		}, nil
	}

	return &common.WriteResult{
		Success:  false,
		RecordId: i.ID,
		Errors:   datautils.ToAnySlice(i.Errors),
		Data:     nil,
	}, nil
}

func (i UpdateItem) ToWriteResult() (*common.WriteResult, error) {
	success := len(i.Errors) == 0

	if success {
		data, err := common.RecordDataToMap(i)
		if err != nil {
			return nil, err
		}

		return &common.WriteResult{
			Success:  true,
			RecordId: i.ID,
			Errors:   nil,
			Data:     data,
		}, nil
	}

	return &common.WriteResult{
		Success:  false,
		RecordId: i.ID,
		Errors:   datautils.ToAnySlice(i.Errors),
		Data:     nil,
	}, nil
}

func createUnprocessableItem(payloadItem PayloadItem) *common.WriteResult {
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
	}
}
