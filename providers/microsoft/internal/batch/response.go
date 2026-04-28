package batch

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/codec"
	"github.com/amp-labs/connectors/internal/httpkit"
)

// Result contains the outcome of a batch execution.
//
// It preserves both structured results and raw responses for debugging
// and advanced inspection.
type Result[B any] struct {
	// Responses contains all successful (2xx) results keyed by RequestID.
	Responses map[RequestID]Envelope[B]

	// Errors contains all failed results keyed by RequestID.
	//
	// Includes:
	//   - API errors (non-2xx responses)
	//   - connector-level failures (Status = -1)
	Errors map[RequestID]Envelope[error]
}

const connectorErrorStatus = -1

// Envelope represents the outcome of a single request within a batch.
type Envelope[B any] struct {
	// HTTP status code returned by the API.
	// For connector-level failures, Status is set to connectorErrorStatus.
	Status int

	// Data contains:
	//   - decoded response body (for success)
	//   - error (for failures)
	Data B
}

// GetInOrder returns successful response bodies in the order of provided IDs.
// Missing or failed IDs will not be returned.
func (r Result[B]) GetInOrder(requestIdentifiers []RequestID) []B {
	result := make([]B, 0, len(requestIdentifiers))
	for _, identifier := range requestIdentifiers {
		result = append(result, r.Responses[identifier].Data)
	}

	return result
}

func (r Result[B]) JoinedErr() error {
	var err error
	for _, e := range r.Errors {
		err = errors.Join(err, e.Data)
	}

	return err
}

// storeResponseBody classifies a single API response into success or failure.
//
// Non-2xx responses are converted into structured errors and stored in Errors.
// Successful responses are stored in Responses.
func (r Result[B]) storeResponseBody(wrapper responseWrapper[B]) { // nolint:unparam
	item := wrapper.Data
	if !httpkit.Status2xx(item.Status) {
		// In case of an error, no response body will be included.
		// Unlikely and not a big deal.
		data, _ := json.Marshal(wrapper.Raw) // nolint:errchkjson
		r.Errors[item.RequestID] = Envelope[error]{
			Status: item.Status,
			Data: common.NewHTTPError(
				item.Status, data, item.getHeaders(), ErrBatchResponse,
			),
		}

		return
	}

	r.Responses[item.RequestID] = Envelope[B]{
		Status: item.Status,
		Data:   item.Body,
	}
}

// responses models the top-level Graph batch response structure.
type responses[B any] struct {
	Responses []responseWrapper[B] `json:"responses"`
}

// responseWrapper preserves both raw JSON and decoded structure.
//
// This enables:
//   - lossless error reporting
//   - typed access to successful responses
type responseWrapper[B any] = codec.RawJSON[response[B]]

// response is a template Microsoft Graph API formats its data.
// The Body various based on the Request Endpoint and the status of the response.
type response[B any] struct {
	RequestID RequestID      `json:"id"`
	Status    int            `json:"status"`
	Headers   map[string]any `json:"headers"`
	Body      B              `json:"body"`
}

func (b response[B]) getHeaders() common.Headers {
	headers := make(common.Headers, 0)
	for key, value := range b.Headers {
		headers = append(headers, common.Header{
			Key:   key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	return headers
}
