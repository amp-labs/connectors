package common

import (
	"errors"
	"net/http"
)

// BatchReadResult is the result of GetRecordsByIds: the fetched records and the ids that failed.
type BatchReadResult struct {
	// Rows are the fetched records.
	Rows []ReadResultRow `json:"rows"`

	// Failures are the requested ids that could not be fetched.
	Failures []RecordFailure `json:"failures,omitempty"`
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
	case ErrorClassRateLimited,
		ErrorClassProvider5xx,
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

// AddFailure adds a failed id to the result.
func (r *BatchReadResult) AddFailure(recordId string, err error) {
	r.Failures = append(r.Failures, NewRecordFailure(recordId, err))
}

// HasFailures reports whether any requested id failed to fetch.
func (r *BatchReadResult) HasFailures() bool {
	return r != nil && len(r.Failures) > 0
}
