package batch

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// BatchWrite performs a HubSpot batch create or update request.
//
// The request body always includes an "inputs" array of record payloads,
// and the response contains a "results" array aligned by index.
//
// HubSpot may return 400 (Bad Request) or 409 (Conflict) when record-level
// validation fails — these are treated as soft issues (non-fatal responses)
// and are parsed into a structured BatchWriteResult rather than raised as errors.
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

// parseBulkIssue converts structured HubSpot error responses (400, 409)
// into a BatchWriteResult that marks all records as failed.
//
// HubSpot may return more error messages than there are input records.
// For example, if two records each contain three invalid fields, the response
// may contain six validation errors in total. Because these messages lack
// per-record identifiers, the errors cannot be matched back to individual payloads.
// In such cases, all errors are returned at the BatchWriteResult's top level.
func parseBulkIssue(payload *Payload, rsp *common.JSONHTTPResponse) (*common.BatchWriteResult, error) {
	response, err := common.UnmarshalJSON[IssueResponse](rsp)
	if err != nil {
		return nil, err
	}

	var failures []any
	if len(response.Errors) == 0 && response.Message != "" {
		// Include the general message only when no per-record errors exist.
		failures = []any{response.Message}
	} else {
		failures = datautils.ToAnySlice(response.Errors)
	}

	totalNumRecords := len(payload.Items)

	return common.NewBatchWriteResultFailed(nil, totalNumRecords, failures)
}

// parseBulkResponse handles successful (2xx) HubSpot batch responses.
// It maps each response item back to its corresponding payload record,
// producing a per-record WriteResult when possible.
//
// For create operations, HubSpot preserves payload order — the Nth response
// corresponds to the Nth payload. For updates, results are matched by record ID.
func parseBulkResponse(
	params *common.BatchWriteParam, payload *Payload, rsp *common.JSONHTTPResponse,
) (*common.BatchWriteResult, error) {
	response, err := common.UnmarshalJSON[Response](rsp)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return common.NewBatchWriteResultFailed(
			nil, len(payload.Items), []any{common.ErrEmptyJSONHTTPResponse})
	}

	// == Create == //
	if params.IsCreate() {
		// Response order matches payload order.
		return common.ParseBatchWrite(
			payload.Items,
			func(index int, payloadItem PayloadItem) *ResponseItem {
				return &response.Results[index]
			},
			func(_ PayloadItem, respItem *ResponseItem) (*common.WriteResult, error) {
				if respItem == nil {
					return createUnprocessableItem(""), nil
				}

				return respItem.ToWriteResult()
			},
			datautils.ToAnySlice(response.Errors),
		)
	}

	// == UPDATE == //
	// Build a lookup table keyed by record ID.
	items := response.GetItemsMap()

	return common.ParseBatchWrite(
		payload.Items,
		func(_ int, payloadItem PayloadItem) *ResponseItem {
			// Each record must have an id when performing Bulk Update.
			return items[payloadItem.ID]
		},
		func(payloadItem PayloadItem, respItem *ResponseItem) (*common.WriteResult, error) {
			if respItem == nil {
				// No matching response, but we still know which record failed.
				return createUnprocessableItem(payloadItem.ID), nil
			}

			return respItem.ToWriteResult()
		},
		datautils.ToAnySlice(response.Errors),
	)
}

func (a *Adapter) buildBatchWriteURL(params *common.BatchWriteParam) (*urlbuilder.URL, error) {
	if params.IsCreate() {
		return a.getCreateURL(params.ObjectName)
	}

	if params.IsUpdate() {
		return a.getUpdateURL(params.ObjectName)
	}

	return nil, common.ErrUnsupportedBatchWriteType
}

func buildBatchWritePayload(params *common.BatchWriteParam) (*Payload, error) {
	payloadItems := make([]PayloadItem, len(params.Batch))

	for index, batchItem := range params.Batch {
		record, err := batchItem.GetRecord()
		if err != nil {
			return nil, err
		}

		item, err := NewPayloadItem(record, batchItem.Associations)
		if err != nil {
			return nil, err
		}

		payloadItems[index] = *item
	}

	return &Payload{Items: payloadItems}, nil
}

func createUnprocessableItem(identifier string) *common.WriteResult {
	return &common.WriteResult{
		Success:  false, // not processed
		RecordId: identifier,
		Errors: []any{
			common.ErrBatchUnprocessedRecord,
		},
		Data: nil,
	}
}

// Payload represents the HubSpot batch request body.
type Payload struct {
	Items []PayloadItem `json:"inputs"`
}

// PayloadItem represents a single item in the API payload.
// Hubspot's payload is identical to what client supplies to the connector.
// This is an alias.
type PayloadItem struct {
	ID           string        `json:"id,omitempty"`
	Properties   common.Record `json:"properties"`
	Associations any           `json:"associations,omitempty"`
}

func NewPayloadItem(record common.Record, associations any) (*PayloadItem, error) {
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

	// Hubspot will complain about unexpected fields, must cleanup
	delete(properties, "id")

	return &PayloadItem{
		ID:           identifier,
		Properties:   common.Record(properties),
		Associations: associations,
	}, nil
}

// Response models a HubSpot batch success response.
type Response struct {
	CompletedAt time.Time      `json:"completedAt"`
	Status      string         `json:"status"`
	StartedAt   time.Time      `json:"startedAt"`
	Results     []ResponseItem `json:"results"`
	Errors      []Issue        `json:"errors"`
}

func (r Response) GetItemsMap() map[string]*ResponseItem {
	mapping := make(map[string]*ResponseItem)

	for _, item := range r.Results {
		mapping[item.ID] = &item
	}

	return mapping
}

type ResponseItem struct {
	ID         string    `json:"id"`
	Properties any       `json:"properties"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Archived   bool      `json:"archived"`
	URL        string    `json:"url"`
}

func (i ResponseItem) ToWriteResult() (*common.WriteResult, error) {
	data, err := common.RecordDataToMap(i.Properties)
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

// IssueResponse models HubSpot’s structured error response for 4xx cases.
type IssueResponse struct {
	Status        string  `json:"status,omitempty"`
	Message       string  `json:"message,omitempty"`
	CorrelationId string  `json:"correlationId,omitempty"`
	Errors        []Issue `json:"errors,omitempty"`
	Category      string  `json:"category,omitempty"`
}

// Issue represents a single HubSpot error entry.
// Its structure varies by error type, but typically includes "message" and "context" fields.
type Issue any
