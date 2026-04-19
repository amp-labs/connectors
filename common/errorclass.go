// nolint:revive,godoclint
package common

import (
	"errors"
	"net/http"
	"strings"
)

// ErrorClass is a stable, coarse category for an error. It is designed to be
// consumed by observability tooling (Sentry tags, alert routing, dashboards)
// and by callers that need to react to broad classes of failure without
// coupling to specific sentinel values.
//
// New classes should be added here rather than invented at a call site so that
// downstream systems (Sentry saved searches, alert rules) can stay in sync
// with the set of possible values.
type ErrorClass string

const (
	// ErrorClassNone is used when there is no error (err == nil).
	ErrorClassNone ErrorClass = "none"

	// ErrorClassUnknown means no classifier matched. Treat as "needs triage".
	ErrorClassUnknown ErrorClass = "unknown"

	// ErrorClassAuthInvalidated — credentials rejected; needs reconnect.
	ErrorClassAuthInvalidated ErrorClass = "auth_invalidated"

	// ErrorClassForbidden — authenticated but not authorized for this resource.
	ErrorClassForbidden ErrorClass = "forbidden"

	// ErrorClassAPIDisabled — the provider-side feature/API is disabled for
	// this customer's instance (e.g. NetSuite REST Web Services off).
	ErrorClassAPIDisabled ErrorClass = "api_disabled"

	// ErrorClassCursorGone — pagination cursor expired or invalidated.
	ErrorClassCursorGone ErrorClass = "cursor_gone"

	// ErrorClassRateLimited — 429 or provider-specific quota error.
	ErrorClassRateLimited ErrorClass = "rate_limited"

	// ErrorClassBadRequest — 4xx caller error that won't succeed on retry.
	ErrorClassBadRequest ErrorClass = "bad_request"

	// ErrorClassSchemaDriftField — provider rejected a known field name
	// (custom field removed, standard field gated by FLS, typo, etc.).
	ErrorClassSchemaDriftField ErrorClass = "schema_drift_missing_field"

	// ErrorClassSchemaDriftObject — provider rejected a known object name.
	ErrorClassSchemaDriftObject ErrorClass = "schema_drift_missing_object"

	// ErrorClassProviderMigration — the provider is transiently unavailable
	// due to an internal migration (e.g. HubSpot hub data-center move).
	// Usually self-heals within hours.
	ErrorClassProviderMigration ErrorClass = "provider_migration"

	// ErrorClassProvider5xx — 5xx server error on the provider side. Retryable.
	ErrorClassProvider5xx ErrorClass = "provider_5xx"

	// ErrorClassRetryable — generic retryable error where a finer class isn't known.
	ErrorClassRetryable ErrorClass = "retryable"
)

// Classifier is implemented by error types that know their own ErrorClass.
// Callers should prefer `ClassOf(err)` over direct `errors.As` checks because
// ClassOf walks the error chain and applies fallbacks.
type Classifier interface {
	ErrorClass() ErrorClass
}

// classedError wraps a real errors.New sentinel with classification metadata.
// The actual error is stored inside and accessible via Unwrap(), preserving the
// full error tree. Classification is purely additive — it never changes the
// error's message or identity.
//
// errors.Is(wrappedErr, ErrAccessToken) works by pointer identity on the
// *classedError, exactly as it did when ErrAccessToken was errors.New directly.
type classedError struct {
	err   error      // the actual error (errors.New)
	class ErrorClass // classification metadata
}

func (e *classedError) Error() string          { return e.err.Error() }
func (e *classedError) Unwrap() error          { return e.err }
func (e *classedError) ErrorClass() ErrorClass { return e.class }

// newClassedErr constructs a classified sentinel error. The actual error
// (errors.New) is stored inside and accessible via Unwrap(). The returned
// pointer is stable for the life of the process — assign to a package-level
// var and use with errors.Is / errors.As as you would any sentinel.
func newClassedErr(msg string, class ErrorClass) *classedError {
	return &classedError{err: errors.New(msg), class: class} //nolint:err113
}

// ClassOf returns the ErrorClass for an error, walking the error chain.
//
// Resolution order:
//  1. Any error in the chain implementing Classifier wins (sentinels declared
//     with newClassedErr implement it automatically).
//  2. String-content heuristics for cases where Temporal or another layer has
//     flattened the chain into a raw string and lost the typed sentinel.
//
// Returns ErrorClassNone for nil; never panics.
func ClassOf(err error) ErrorClass {
	if err == nil {
		return ErrorClassNone
	}

	var c Classifier
	if errors.As(err, &c) {
		return c.ErrorClass()
	}

	return classOfMessage(err.Error())
}

// classOfMessage is the string-content fallback. Patterns derived from
// production workflow-error samples. Each entry is a class that we can't
// reliably derive from the typed chain today.
func classOfMessage(raw string) ErrorClass {
	msg := strings.ToLower(raw)

	switch {
	case strings.Contains(msg, "no such column"),
		strings.Contains(msg, "didn't understand relationship"):
		return ErrorClassSchemaDriftField
	case strings.Contains(msg, "resource not found for the segment"),
		strings.Contains(msg, "could not find a property named"):
		return ErrorClassSchemaDriftObject
	case strings.Contains(msg, "currently being migrated"):
		return ErrorClassProviderMigration
	case strings.Contains(msg, "access token invalid"),
		strings.Contains(msg, "credentials have been marked as invalid"),
		strings.Contains(msg, "this user is locked"):
		return ErrorClassAuthInvalidated
	case strings.Contains(msg, "feature_disabled"),
		strings.Contains(msg, "api has not been used in project"):
		return ErrorClassAPIDisabled
	case strings.Contains(msg, "cursor has expired"):
		return ErrorClassCursorGone
	case strings.Contains(msg, "request header fields too large"),
		strings.Contains(msg, "precondition check failed"),
		strings.Contains(msg, "missing secret"):
		return ErrorClassBadRequest
	case strings.Contains(msg, "http status 5"),
		strings.Contains(msg, "internal server error"),
		strings.Contains(msg, "unexpected server error"):
		return ErrorClassProvider5xx
	}

	return ErrorClassUnknown
}

// classOfHTTPStatus maps an HTTP status code to an ErrorClass. Used by
// HTTPError.ErrorClass as the first signal — more specific than the wrapped
// sentinel in many cases (e.g. a 477 from HubSpot is "provider_migration"
// regardless of what sentinel it wraps).
// hubspotMigrationStatus is the non-standard HTTP 477 status code HubSpot
// returns when a hub is being migrated between data hosting locations.
const hubspotMigrationStatus = 477

func classOfHTTPStatus(status int) (ErrorClass, bool) {
	switch status {
	case http.StatusUnauthorized:
		return ErrorClassAuthInvalidated, true
	case http.StatusForbidden:
		return ErrorClassForbidden, true
	case http.StatusTooManyRequests:
		return ErrorClassRateLimited, true
	case hubspotMigrationStatus:
		return ErrorClassProviderMigration, true
	}

	if status >= 500 && status < 600 {
		return ErrorClassProvider5xx, true
	}

	return "", false
}
