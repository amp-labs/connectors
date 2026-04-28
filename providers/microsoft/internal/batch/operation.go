package batch

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// ErrBatchResponse represents a failure returned by the Microsoft Graph Batch API
// for an individual request within a batch.
var ErrBatchResponse = errors.New("API error in batch response")

// Execute performs a Microsoft Graph JSON batch request.
//
// It aggregates multiple logical HTTP requests into one call to the `$batch` endpoint:
// https://learn.microsoft.com/en-us/graph/json-batching
//
// Core Contract:
//   - Every RequestID provided is deterministically mapped to exactly one outcome:
//   - success → stored in Result.Responses
//   - failure → stored in Result.Errors
//   - Failures may originate from:
//   - API-level errors (non-2xx responses per request)
//   - connector-level errors (transport, serialization, or chunk failure)
//
// Behavior:
//   - Requests are automatically chunked to respect Graph API limits (max 20 per batch).
//   - Absolute URLs are normalized to relative paths as required by Graph.
//   - Each batch is executed independently; failures in one batch do not prevent others.
//
// Error Handling:
//   - This method does NOT return a combined error.
//   - Callers must inspect Result.Errors to handle failures.
//   - This design ensures full visibility into partial success scenarios.
//
// Generics:
//   - `B` defines the expected response body type.
//   - The caller is responsible for ensuring compatibility with each request.
//
// Guarantees:
//   - Each RequestID appears in either Responses or Errors (never both, never neither).
//   - Order is not preserved unless explicitly reconstructed via List().
//
// Example:
//
//	req := new(BatchRequest).
//	    WithRequest("1", "GET", url1, nil, nil).
//	    WithRequest("2", "POST", url2, body, nil)
//
//	result := Execute[MyResponse](ctx, strategy, req)
//
//	success := result.Responses["1"]
//	failure := result.Errors["2"]
func Execute[B any](ctx context.Context, strategy *Strategy, params *Params) *Result[B] {
	bundle := &Result[B]{
		Responses: make(map[RequestID]Envelope[B]),
		Errors:    make(map[RequestID]Envelope[error]),
	}

	// Chunk the list of payloads into consumable sizes, otherwise API will reject large number of requests.
	// Save the output into the bundle.Raw and bundle.Response.
	// In case of an error the bundle.Raw and bundle.Errors are populated.
	for _, payloads := range params.chunkPayloads() {
		res, err := strategy.performRequest(ctx, payloads)
		if err != nil {
			for _, payload := range payloads {
				bundle.Errors[payload.RequestID] = Envelope[error]{
					Status: connectorErrorStatus,
					Data:   fmt.Errorf("batch request failed: %w", err),
				}
			}

			// Go to the next chunk request.
			continue
		}

		apiResponse, err := common.UnmarshalJSON[responses[B]](res)
		if err != nil {
			for _, payload := range payloads {
				bundle.Errors[payload.RequestID] = Envelope[error]{
					Status: connectorErrorStatus,
					Data:   fmt.Errorf("failed to unmarshal batch response: %w", err),
				}
			}

			// Parsing output failed.
			continue
		}

		// Sort every response into either success or failure.
		for _, wrapper := range apiResponse.Responses {
			bundle.storeResponseBody(wrapper)
		}
	}

	return bundle
}

func (s Strategy) performRequest(ctx context.Context, payloads []*payloadRequest) (*common.JSONHTTPResponse, error) {
	url, err := s.getBatchURL()
	if err != nil {
		return nil, err
	}

	for _, payload := range payloads {
		payload.RelativeURL, _ = strings.CutPrefix(payload.RelativeURL, s.getVersionedRootURL())
	}

	return s.client.Post(ctx, url.String(), bundledPayload{
		Requests: payloads,
	})
}
