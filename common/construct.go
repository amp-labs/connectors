// nolint:revive,godoclint
package common

import (
	"errors"

	"github.com/amp-labs/connectors/internal/goutils"
)

var ErrNumWriteResultExceedsTotalRecords = errors.New(
	"number of batch WriteResult entries exceeds total number of payload Records",
)

// NewBatchWriteResult constructs a new BatchWriteResult summarizing the outcome
// of a batch write operation. It validates input consistency, computes
// success/failure counts, and derives the BatchStatus based on those counts.
//
// Rules and assumptions:
//
//   - len(results) must not exceed totalNumRecords.
//     If it does, the constructor returns a joined error containing
//     ErrInvalidImplementation and ErrNumWriteResultExceedsTotalRecords.
//
//   - successCounter < 0 triggers automatic counting of successes from the
//     results slice. This is a convenience for callers that did not precompute
//     successes. However, providing an explicit success count is preferred.
//
//   - successCounter cannot exceed totalNumRecords.
//     If it does, the counter is recomputed defensively.
//
//   - failureCounter is computed as totalNumRecords - successCounter.
//
//   - A BatchWriteResult always satisfies:
//     SuccessCount + FailureCount == totalNumRecords
//     and
//     len(Results) ≤ totalNumRecords.
//
//   - unmatchedErrors represent provider responses that could not be associated
//     with specific payload items — for example, schema validation issues or
//     general API errors returned alongside per-record results.
//
// Constructors may return an error to signal invalid or inconsistent usage
// rather than to represent runtime provider failures.
func NewBatchWriteResult(
	results []WriteResult, successCounter, totalNumRecords int, unmatchedErrors []any,
) (*BatchWriteResult, error) {
	if len(results) > totalNumRecords {
		return nil, errors.Join(ErrInvalidImplementation, ErrNumWriteResultExceedsTotalRecords)
	}

	if successCounter < 0 || successCounter > totalNumRecords {
		successCounter = countSuccesses(results)
	}

	failureCounter := totalNumRecords - successCounter

	return &BatchWriteResult{
		Status:       newBatchStatus(successCounter, failureCounter, totalNumRecords),
		Errors:       unmatchedErrors,
		Results:      results,
		SuccessCount: successCounter,
		FailureCount: failureCounter,
	}, nil
}

// NewBatchWriteResultFailed constructs a BatchWriteResult representing a fully
// failed batch operation. It assumes zero successful records and marks the
// BatchStatus as failure. The constructor still validates that the number of
// WriteResult entries does not exceed totalNumRecords.
//
// unmatchedErrors may include provider-level issues explaining the failure that cannot be tied to specific records.
func NewBatchWriteResultFailed(
	results []WriteResult, totalNumRecords int, unmatchedErrors []any,
) (*BatchWriteResult, error) {
	if len(results) > totalNumRecords {
		return nil, errors.Join(ErrInvalidImplementation, ErrNumWriteResultExceedsTotalRecords)
	}

	return &BatchWriteResult{
		Status:       newBatchStatus(0, totalNumRecords, totalNumRecords),
		Errors:       unmatchedErrors,
		Results:      results,
		SuccessCount: 0,
		FailureCount: totalNumRecords,
	}, nil
}

func newBatchStatus(successCounter, failureCounter, total int) BatchStatus {
	switch {
	case successCounter == total:
		return BatchStatusSuccess
	case failureCounter == total:
		// Every single record failed.
		return BatchStatusFailure
	default:
		// Some failed, some succeeded.
		return BatchStatusPartial
	}
}

// BatchWriteResponseMatcher matches a single payload item from the request
// to its corresponding provider response item.
//
// Implementations may match items by inspecting payload data or by using the index,
// depending on how the provider structures its response. For providers that return
// responses in the same order as the request, the index allows a straightforward positional lookup.
//
// The function must be deterministic and fast — given a payload item (and its index),
// it should return the corresponding response item or nil if none exists.
type BatchWriteResponseMatcher[P, R any] func(index int, payloadItem P) *R

// BatchWriteResponseTransformer converts a payload item and its matched provider response
// into a standardized WriteResult.
//
// Responsibilities of the transformer:
//   - Determine success vs failure for the item (WriteResult.Success).
//   - Populate WriteResult.RecordId when the provider returns an identifier for the affected record.
//   - Attach errors describing what went wrong at WriteResult.Errors
//   - In case of successful create/update set the response item to WriteResult.Data.
//
// Type parameters:
//
//	P — payload item type.
//	R — provider response item type.
type BatchWriteResponseTransformer[P, R any] func(payloadItem P, respItem *R) (*WriteResult, error)

// ErrBatchUnprocessedRecord is returned when a record was skipped or not
// processed due to failures elsewhere in the batch.
var ErrBatchUnprocessedRecord = errors.New("record was not processed due to other records failures")

// ParseBatchWrite converts a provider's overall batch response into a consolidated BatchWriteResult.
//
// A Record is a domain-agnostic map/struct that models fields you want to persist.
// An Item is the wire representation (payload item) the provider expects for a single Record.
// Connectors must make an explicit mapping Record -> Item (payload item) when constructing the overall payload.
// When responses arrive, connectors must match each payload item to its response item and produce a WriteResult.
// This separation ensures consistent handling of ID mapping, reference IDs, partial successes, and errors.
//
// --------------------------
// The function is intentionally generic:
//
//	P is the payload item type (what was sent),
//	R is the provider response item type (what was received).
//
// Arguments:
// payloadItems		- list of items that are part of payload to create/update each record.
// responseMatcher	- list of items that are part of payload to create/update each record.
// responseToResult	- a transformer that converts a matched item pair (payload P, response R) into a WriteResult.
// unmatchedErrors	– top-level errors not tied to individual records,
//
//	such as validation failures detected before response matching.
func ParseBatchWrite[P, R any](
	payloadItems []P,
	responseMatcher BatchWriteResponseMatcher[P, R],
	responseToResult BatchWriteResponseTransformer[P, R],
	unmatchedErrors []any,
) (*BatchWriteResult, error) {
	var (
		totalNumRecords = len(payloadItems)
		results         = make([]WriteResult, 0, totalNumRecords)
		successCounter  = 0
	)

	for index, record := range payloadItems {
		response, err := invokeResponseMatcher(responseMatcher, index, record)
		if err != nil {
			// Index out of bounds is downgraded from panic to error inside the invoker.
			return nil, err
		}

		result, err := responseToResult(record, response)
		if err != nil {
			unmatchedErrors = append(unmatchedErrors, err)

			// Record cannot be added into the list of results ([]WriteResult).
			continue
		}

		// The result added could be either successful or failed.
		// Each has a mixed status. Therefore, we keep track of successes and failures separately.
		results = append(results, *result)

		if result.Success {
			successCounter += 1
		}
	}

	return NewBatchWriteResult(results, successCounter, totalNumRecords, unmatchedErrors)
}

func countSuccesses(results []WriteResult) int {
	count := 0

	for _, result := range results {
		if result.Success {
			count += 1
		}
	}

	return count
}

// invokeResponseMatcher safely executes the provided response matcher function.
// It guards against panics—such as index out-of-range or unexpected nil access—
// converting them into regular errors instead of crashing.
//
// This acts as a safety net for connector implementors who may have used
// the index incorrectly when matching payload and response items.
func invokeResponseMatcher[P, R any](
	responseMatcher BatchWriteResponseMatcher[P, R],
	index int, record P,
) (responseItem *R, err error) {
	defer goutils.PanicRecovery(func(cause error) {
		err = errors.Join(ErrInvalidImplementation, cause)
		responseItem = nil
	})

	return responseMatcher(index, record), nil
}
