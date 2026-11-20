package common

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

// BatchReadResult is the result of GetRecordsByIds: the fetched records and the ids that failed.
//
// A connector that fetches each id with its own request calls AddSuccess or AddFailure as each fetch
// finishes, then returns Err alongside the result. AddSuccess and AddFailure are safe to call from
// concurrent fetches; rows and failures are kept in the order the calls were made.
type BatchReadResult struct {
	// Rows are the fetched records.
	Rows []ReadResultRow `json:"rows"`

	// Failures are the requested ids that could not be fetched.
	Failures []RecordFailure `json:"failures,omitempty"`

	mutex sync.Mutex
}

// RecordFailure is a single requested id that could not be fetched.
type RecordFailure struct {
	// RecordId is the id as passed to GetRecordsByIds.
	RecordId string `json:"recordId"`

	// Err is the underlying error. It is not serialized.
	Err error `json:"-"`

	// Reason tells the caller whether the record is worth retrying.
	Reason FailureReason `json:"reason"`
}

// FailureReason categorizes a per-record failure for retry decisions.
type FailureReason string

const (
	// FailureReasonUnknown means the failure could not be categorized.
	FailureReasonUnknown FailureReason = "unknown"

	// FailureReasonNotFound means the record does not exist.
	FailureReasonNotFound FailureReason = "not_found"

	// FailureReasonRateLimited means the provider rate-limited the request; the fetch may succeed
	// after backing off.
	FailureReasonRateLimited FailureReason = "rate_limited"

	// FailureReasonTransient means the fetch may succeed on retry.
	FailureReasonTransient FailureReason = "transient"

	// FailureReasonPermanent means the fetch will fail again on retry.
	FailureReasonPermanent FailureReason = "permanent"
)

// NewRecordFailure builds a RecordFailure with its Reason derived from err.
func NewRecordFailure(recordId string, err error) RecordFailure {
	return RecordFailure{
		RecordId: recordId,
		Err:      err,
		Reason:   FailureReasonOf(err),
	}
}

// FailureReasonOf maps an error to a FailureReason.
func FailureReasonOf(err error) FailureReason {
	if err == nil {
		return FailureReasonUnknown
	}

	if isNotFoundErr(err) {
		return FailureReasonNotFound
	}

	switch ClassOf(err) {
	case ErrorClassRateLimited:
		return FailureReasonRateLimited

	case ErrorClassProvider5xx,
		ErrorClassProviderMigration,
		ErrorClassRetryable:
		return FailureReasonTransient

	case ErrorClassAuthInvalidated,
		ErrorClassForbidden,
		ErrorClassAPIDisabled,
		ErrorClassBadRequest,
		ErrorClassCursorGone,
		ErrorClassSchemaDriftField,
		ErrorClassSchemaDriftObject,
		ErrorClassProvider5xxPermanent:
		return FailureReasonPermanent

	case ErrorClassNone, ErrorClassUnknown:
		return FailureReasonUnknown

	default:
		return FailureReasonUnknown
	}
}

// isNotFoundErr reports whether err means the record does not exist.
func isNotFoundErr(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}

	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		return httpErr.Status == http.StatusNotFound
	}

	return false
}

// NewBatchReadResult creates a BatchReadResult with the given rows and no failures.
func NewBatchReadResult(rows []ReadResultRow) *BatchReadResult {
	return &BatchReadResult{Rows: rows}
}

// ReportMissingIds adds a not-found failure for each requested id missing from Rows.
func (r *BatchReadResult) ReportMissingIds(requestedIds []string) {
	if r == nil {
		return
	}

	returned := make(map[string]bool, len(r.Rows))
	for _, row := range r.Rows {
		returned[row.Id] = true
	}

	reported := make(map[string]bool, len(requestedIds))

	for _, id := range requestedIds {
		if returned[id] || reported[id] {
			continue
		}

		reported[id] = true

		r.AddFailure(id, fmt.Errorf("%w: %s not returned by the batch request", ErrNotFound, id))
	}
}

// AddSuccess adds the rows fetched for one requested id. It is safe to call from concurrent fetches.
func (r *BatchReadResult) AddSuccess(rows ...ReadResultRow) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Rows = append(r.Rows, rows...)
}

// AddFailure adds a failed id to the result, with the reason derived from err. It is safe to call
// from concurrent fetches.
func (r *BatchReadResult) AddFailure(recordId string, err error) {
	failure := NewRecordFailure(recordId, err)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.Failures = append(r.Failures, failure)
}

// Err joins the errors of the failures that are not not-found, for the connector to return alongside
// the result. It is nil when every failure is not-found, so a missing record does not fail the call.
// Call it after every fetch has finished.
func (r *BatchReadResult) Err() error {
	if r == nil {
		return nil
	}

	var errs []error

	for _, failure := range r.Failures {
		if failure.Reason != FailureReasonNotFound {
			errs = append(errs, fmt.Errorf("record %s: %w", failure.RecordId, failure.Err))
		}
	}

	return errors.Join(errs...)
}

// HasFailures reports whether any requested id failed to fetch.
func (r *BatchReadResult) HasFailures() bool {
	return r != nil && len(r.Failures) > 0
}

// ReportFetchErrors records the per-id errors of a fetch that requests each id separately. errs is
// keyed by the id's position in recordIds. Every error is added as a failure, in request order, with
// its reason. The errors that are not not-found are also returned, joined, so the caller can return
// them alongside the result.
//
// Deprecated: call AddSuccess or AddFailure as each id's fetch finishes, then return Err.
func (r *BatchReadResult) ReportFetchErrors(recordIds []string, errs map[int]error) error {
	var other []error

	for idx, recordId := range recordIds {
		err, ok := errs[idx]
		if !ok {
			continue
		}

		r.AddFailure(recordId, err)

		if !isNotFoundErr(err) {
			other = append(other, fmt.Errorf("record %s: %w", recordId, err))
		}
	}

	return errors.Join(other...)
}
