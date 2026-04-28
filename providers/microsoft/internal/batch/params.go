package batch

import (
	"slices"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

// RequestID uniquely identifies a request within a batch.
//
// It is the sole mechanism used to correlate requests and responses.
//
// Design Notes:
//   - Strongly typed to prevent accidental mixing with arbitrary strings.
//   - Fully controlled by the caller (no imposed format).
//   - Encourages explicit mapping logic at call sites.
type RequestID string

// Params collects individual requests that will be executed as a single batch.
//
// It acts as a builder for batch execution:
//   - Requests are added via WithRequest.
//   - Each request must have a unique RequestID.
//
// Internal behavior such as chunking (to satisfy API limits) is handled automatically
// and is not exposed to the caller.
type Params struct {
	payloads []*payloadRequest
}

// WithRequest adds a request to the batch.
//
// Requirements:
//   - requestIdentifier must be unique within the batch.
//   - URL must be a valid Graph endpoint.
//
// Notes:
//   - The identifier is the only mechanism used to correlate responses.
func (p *Params) WithRequest(
	requestIdentifier RequestID,
	method string,
	url *urlbuilder.URL,
	body any,
	headers map[string]any,
) *Params {
	if p.payloads == nil {
		p.payloads = make([]*payloadRequest, 0)
	}

	p.payloads = append(p.payloads, &payloadRequest{
		RequestID:   requestIdentifier,
		Method:      method,
		RelativeURL: url.String(),
		Body:        body,
		Headers:     headers,
	})

	return p
}

// chunkPayloads splits payloads into batches respecting the Graph API limit.
// https://learn.microsoft.com/en-us/graph/json-batching?tabs=http#batch-size-limitations
func (p *Params) chunkPayloads() [][]*payloadRequest {
	const maxBatchSize = 20

	if len(p.payloads) == 0 {
		return nil
	}

	iterator := slices.Chunk(p.payloads, maxBatchSize)

	var batches [][]*payloadRequest
	iterator(func(batch []*payloadRequest) bool {
		batches = append(batches, batch)

		return true // continue iteration
	})

	return batches
}

// bundledPayload is the main payload sent to "/$batch" endpoint.
type bundledPayload struct {
	Requests []*payloadRequest `json:"requests"`
}

// payloadRequest represents a single request entry in the Graph batch payload.
type payloadRequest struct {
	RequestID   RequestID      `json:"id"`
	Method      string         `json:"method"`
	RelativeURL string         `json:"url"`
	Body        any            `json:"body,omitempty"`
	Headers     map[string]any `json:"headers,omitempty"`
}
