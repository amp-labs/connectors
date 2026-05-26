package batch

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

// ReadParams describes a batch read request.
//
// ObjectName identifies the HubSpot object type to read, and Identifiers contains the object IDs to fetch.
type ReadParams struct {
	ObjectName  common.ObjectName
	Identifiers []string
}

// ReadResult contains the records and errors returned from a batch read.
//
// Records includes all successfully decoded results returned by the API into type [B].
// Errors includes transport, decoding, simultaneously hard errors and API-reported soft errors.
type ReadResult[B any] struct {
	Records []B
	Errors  []error
}

// Read performs a batch read for the requested object type and identifiers.
//
// HubSpot batch read endpoints require identifiers to be sent in chunks, so
// this function splits the request into payloads of the supported size and
// executes them concurrently. Successful results are collected into Records,
// while any transport, decoding, or API soft errors are collected into Errors.
func Read[B any](ctx context.Context, adapter *Adapter, params ReadParams) *ReadResult[B] {
	var (
		numRecords      = len(params.Identifiers)
		responseChannel = make(chan B, numRecords)
		errChannel      = make(chan error, numRecords)
		result          = &ReadResult[B]{
			Records: make([]B, 0, numRecords),
			Errors:  make([]error, 0, numRecords),
		}
	)

	payloadList := newPayloadChunks(params.Identifiers)

	callbacks := make([]simultaneously.Job, len(payloadList))
	for index, payload := range payloadList {
		callbacks[index] = func(ctx context.Context) error {
			readRoutine(ctx, adapter, params, payload, responseChannel, errChannel)

			return nil
		}
	}

	// Wait for all jobs to finish.
	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// All jobs are done, so the channels can be closed safely.
	close(responseChannel)
	close(errChannel)

	for data := range responseChannel {
		result.Records = append(result.Records, data)
	}

	for data := range errChannel {
		result.Errors = append(result.Errors, data)
	}

	return result
}

// readRoutine executes a single batch read request and streams results and
// errors into the provided channels.
//
// API-level errors returned by HubSpot are treated as soft errors and sent to
// errChannel, while successful records are sent to responseChannel.
func readRoutine[B any](ctx context.Context,
	adapter *Adapter,
	params ReadParams,
	payload *readPayload,
	responseChannel chan<- B,
	errChannel chan<- error,
) {
	url, err := adapter.getReadURL(params.ObjectName)
	if err != nil {
		errChannel <- err

		return
	}

	res, err := adapter.Client.Post(ctx, url.String(), payload)
	if err != nil {
		errChannel <- err

		return
	}

	// Parse the response to obtain any API errors.
	apiResponse, err := common.UnmarshalJSON[readResponse[B]](res)
	if err != nil {
		errChannel <- err

		return
	}

	// These are soft errors.
	for _, errorObject := range apiResponse.Errors {
		errChannel <- errors.New(errorObject.Message) // nolint:err113
	}

	for _, data := range apiResponse.Results {
		responseChannel <- data
	}
}

// readPayload is the request body used by the HubSpot batch read endpoint.
type readPayload struct {
	Inputs []readPayloadInput `json:"inputs"`
}

type readPayloadInput struct {
	ID string `json:"id"`
}

// newPayloadChunks splits identifiers into request-sized payloads.
//
// HubSpot batch read endpoints accept a limited number of identifiers per
// request, so this helper chunks the input into batches of that size.
func newPayloadChunks(identifiers []string) []*readPayload {
	// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/batch/get-campaigns
	const maxBatchSize = 50

	if len(identifiers) == 0 {
		return nil
	}

	iterator := slices.Chunk(identifiers, maxBatchSize)

	var batches []*readPayload
	iterator(func(ids []string) bool {
		inputs := make([]readPayloadInput, len(ids))
		for index, id := range ids {
			inputs[index] = readPayloadInput{ID: id}
		}

		batches = append(batches, &readPayload{Inputs: inputs})

		return true // continue iteration
	})

	return batches
}

// readResponse mirrors the HubSpot batch read response.
//
// Results contains successfully decoded records of type [B],
// and Errors contains API errors reported for individual inputs.
type readResponse[B any] struct {
	CompletedAt time.Time `json:"completedAt"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"startedAt"`
	Results     []B       `json:"results"`
	Errors      []struct {
		Status      string `json:"status"`
		Category    string `json:"category"`
		SubCategory string `json:"subCategory"`
		Message     string `json:"message"`
		Context     any    `json:"context"`
	} `json:"errors"`
	NumErrors int `json:"numErrors"`
}
