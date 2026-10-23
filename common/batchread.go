package common

import (
	"errors"
	"net/http"
)

// BatchReadResult is the outcome of a BatchRecordReaderConnector.GetRecordsByIds
// call. It separates the records that were fetched from the ids that could not be,
// so a caller is no longer forced to choose between "all rows" and "one error".
//
// A returned *BatchReadResult is only meaningful when GetRecordsByIds returned a
// nil error. A non-nil error means the call as a whole failed (bad object name,
// malformed request, a batch endpoint that rejected the request outright) and no
// per-id outcome could be determined.
//
// Rows and Failures together do not necessarily cover every requested id: a
// provider that returns fewer records than asked for, without saying which ones
// are missing, produces neither a row nor a failure for the absent ids.
type BatchReadResult struct {
	// Rows are the records that were successfully fetched. Order is
	// provider-dependent and should not be relied upon.
	Rows []ReadResultRow `json:"rows"`

	// Failures are the requested ids that could not be fetched, each with the
	// error that prevented it and a coarse reason the caller can act on.
	Failures []RecordFailure `json:"failures,omitempty"`
}

// RecordFailure is a single requested id that could not be fetched.
type RecordFailure struct {
	// RecordId is the id as it was passed to GetRecordsByIds. For providers with
	// composite ids (e.g. Zoho Mail's "<folderId>/<messageId>") this is the full
	// composite, so the caller can correlate it with what it asked for.
	RecordId string `json:"recordId"`

	// Err is the underlying error. It is not serialized; callers that need to
	// persist a failure should record Reason plus Err.Error() themselves.
	Err error `json:"-"`

	// Reason is the coarse category of the failure, used to decide whether the
	// record is worth retrying.
	Reason FailureReason `json:"reason"`
}

// FailureReason is a coarse, retry-relevant category for a per-record failure.
//
// It is deliberately narrower than ErrorClass: ErrorClass answers "what went
// wrong" for observability, while FailureReason answers the single question a
// caller has to act on — should this id be tried again?
type FailureReason string

const (
	// FailureReasonUnknown means the failure could not be categorized. Callers
	// should treat it conservatively, i.e. like FailureReasonTransient.
	FailureReasonUnknown FailureReason = "unknown"

	// FailureReasonNotFound means the provider positively reported that the
	// record does not exist. Retrying will not help; the record is gone, was
	// never visible to this installation, or the id does not name a record of
	// this object type.
	FailureReasonNotFound FailureReason = "not_found"

	// FailureReasonTransient means the fetch may succeed if tried again —
	// rate limits, provider 5xx, migrations, transport errors.
	FailureReasonTransient FailureReason = "transient"

	// FailureReasonPermanent means the fetch failed for a reason that will
	// recur — revoked auth, missing scope, a rejected field or object name.
	// Retrying the same request is pointless.
	FailureReasonPermanent FailureReason = "permanent"
)

// NewRecordFailure builds a RecordFailure, deriving Reason from err via
// FailureReasonOf. Connectors should prefer this over a struct literal so that
// classification stays consistent across providers; set Reason explicitly only
// where the connector knows something the error itself does not convey.
func NewRecordFailure(recordId string, err error) RecordFailure {
	return RecordFailure{
		RecordId: recordId,
		Err:      err,
		Reason:   FailureReasonOf(err),
	}
}

// FailureReasonOf derives a FailureReason from an error by walking the error
// chain: an explicit not-found first, then the ErrorClass the error carries.
//
// Returns FailureReasonUnknown for a nil error, since a nil error does not
// describe a failure.
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

// isNotFoundErr reports whether err represents a missing record. The JSON HTTP
// client maps some 404s to a *HTTPError rather than ErrNotFound, so the status
// is inspected as well as the sentinel — mirroring what the acculynx and gmail
// connectors already do inline.
func isNotFoundErr(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}

	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		return httpErr.Status == http.StatusNotFound
	}

	return false
}

// NewBatchReadResult wraps successfully fetched rows in a BatchReadResult with
// no failures. It is the common case for providers whose batch endpoint either
// succeeds as a whole or fails as a whole.
func NewBatchReadResult(rows []ReadResultRow) *BatchReadResult {
	return &BatchReadResult{Rows: rows}
}

// AddFailure records a failed id on the result, deriving the reason from err.
func (r *BatchReadResult) AddFailure(recordId string, err error) {
	r.Failures = append(r.Failures, NewRecordFailure(recordId, err))
}

// HasFailures reports whether any requested id failed to fetch.
func (r *BatchReadResult) HasFailures() bool {
	return r != nil && len(r.Failures) > 0
}
