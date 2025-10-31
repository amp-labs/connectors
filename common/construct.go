package common

import "errors"

// NewBatchWriteResult constructs a new BatchWriteResult summarizing the outcome
// of a batch write operation. It calculates success/failure counts and assigns
// the corresponding BatchStatus based on those counts.
func NewBatchWriteResult(
	fatalErrors []any, results []WriteResult,
	successCounter, totalNumRecords int,
) *BatchWriteResult {
	failureCounter := totalNumRecords - successCounter

	return &BatchWriteResult{
		Status:       newBatchStatus(successCounter, failureCounter, totalNumRecords),
		Errors:       fatalErrors,
		Results:      results,
		SuccessCount: successCounter,
		FailureCount: failureCounter,
	}
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
// Implementations may perform the match by inspecting the payload item's data
// or by using its index, depending on how the provider structures its response.
// For providers that return responses in the same order as the request,
// the index argument enables a simple positional lookup.
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
// payloadItems - list of items that are part of payload to create/update each record.
// responseMatcher - list of items that are part of payload to create/update each record.
func ParseBatchWrite[P, R any](
	payloadItems []P,
	responseMatcher BatchWriteResponseMatcher[P, R],
	responseToResult BatchWriteResponseTransformer[P, R],
) (*BatchWriteResult, error) {
	var (
		totalNumRecords = len(payloadItems)
		results         = make([]WriteResult, 0, totalNumRecords)
		fatalErrors     []any
		successCounter  = 0
	)

	for index, record := range payloadItems {
		response := responseMatcher(index, record)

		result, err := responseToResult(record, response)
		if err != nil {
			fatalErrors = append(fatalErrors, err)

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

	return NewBatchWriteResult(fatalErrors, results, successCounter, totalNumRecords), nil
}
