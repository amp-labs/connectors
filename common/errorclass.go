// nolint:revive,godoclint
package common

import (
	"errors"
	"net/http"
	"slices"
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
	// The remedy is provider-side: the connected user's roles, profile or permission
	// sets need changing.
	ErrorClassForbidden ErrorClass = "forbidden"

	// ErrorClassInsufficientScope — the credential itself lacks a scope the request
	// needs. A refinement of ErrorClassForbidden, and worth distinguishing because the
	// remedy is the opposite one: the customer must reconnect granting more scopes, and
	// no amount of provider-side permission fiddling will help.
	ErrorClassInsufficientScope ErrorClass = "insufficient_scope"

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

	// ErrorClassProvider5xxPermanent — a 5xx that a connector has positively
	// identified as permanent. Non-retryable.
	ErrorClassProvider5xxPermanent ErrorClass = "provider_5xx_permanent"

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

// classedMultiError is a classedError that also presents itself as one or more other
// sentinels. It exists so a new, more specific sentinel can be introduced without breaking
// the errors.Is checks written against the broader one it refines.
type classedMultiError struct {
	err   error      // the actual error (errors.New)
	also  []error    // sentinels this error also satisfies
	class ErrorClass // classification metadata
}

func (e *classedMultiError) Error() string          { return e.err.Error() }
func (e *classedMultiError) ErrorClass() ErrorClass { return e.class }

// Unwrap returns the inner error alongside the sentinels this error also satisfies.
// Classification still resolves to e.class because ClassOf takes the outermost Classifier,
// which is this value.
func (e *classedMultiError) Unwrap() []error {
	return append([]error{e.err}, e.also...)
}

// newClassedErrAlso constructs a classified sentinel that additionally satisfies
// errors.Is against each error in `also`.
func newClassedErrAlso(msg string, class ErrorClass, also ...error) *classedMultiError {
	return &classedMultiError{err: errors.New(msg), also: also, class: class} //nolint:err113
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

// messageClassPatterns is the string-content fallback table. Patterns are derived from
// production workflow-error samples and matched against a lowercased message; the first
// entry with a matching pattern wins, so order is significant.
//
// A table rather than a switch so that adding a class is a data change: the switch this
// replaced was already at the cyclomatic complexity ceiling, which made every new class a
// refactor.
var messageClassPatterns = []struct { //nolint:gochecknoglobals
	patterns []string
	class    ErrorClass
}{
	{[]string{"no such column", "didn't understand relationship"}, ErrorClassSchemaDriftField},
	{[]string{"resource not found for the segment", "could not find a property named"}, ErrorClassSchemaDriftObject},
	{[]string{"currently being migrated"}, ErrorClassProviderMigration},
	{
		[]string{"access token invalid", "credentials have been marked as invalid", "this user is locked"},
		ErrorClassAuthInvalidated,
	},
	// "missing_scopes" is HubSpot's error category, "insufficient_scope" is the RFC 6750
	// challenge, and "missing scopes" is our own sentinel's message surviving a flattened
	// chain.
	{[]string{"missing_scopes", "missing scopes", "insufficient_scope"}, ErrorClassInsufficientScope},
	{[]string{"feature_disabled", "api has not been used in project"}, ErrorClassAPIDisabled},
	{[]string{"cursor has expired"}, ErrorClassCursorGone},
	{
		[]string{"request header fields too large", "precondition check failed", "missing secret"},
		ErrorClassBadRequest,
	},
	{[]string{"non-retryable server error"}, ErrorClassProvider5xxPermanent},
	{
		[]string{"http status 5", "internal server error", "unexpected server error"},
		ErrorClassProvider5xx,
	},
}

// classOfMessage is the string-content fallback, for cases where Temporal or another layer
// has flattened the error chain into a raw string and lost the typed sentinel.
func classOfMessage(raw string) ErrorClass {
	msg := strings.ToLower(raw)

	for _, entry := range messageClassPatterns {
		for _, pattern := range entry.patterns {
			if strings.Contains(msg, pattern) {
				return entry.class
			}
		}
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

// statusClassRefinements records, per status-derived class, the classes a connector may
// positively identify as a more specific case of it.
//
// The HTTP status is normally the better signal, which is why HTTPError.ErrorClass prefers
// it. The exception is where the status is simply not capable of carrying the distinction:
// a 403 tells you the provider said no, but not whether the remedy is "reconnect granting
// more scopes" or "change the user's provider-side permissions". Those are opposite
// remedies, so when a connector has read the body and knows which one it is, it wins.
var statusClassRefinements = map[ErrorClass][]ErrorClass{ //nolint:gochecknoglobals
	ErrorClassForbidden: {ErrorClassInsufficientScope},
}

// refinesClass reports whether candidate is a registered, more specific case of base.
func refinesClass(base, candidate ErrorClass) bool {
	if candidate == "" || candidate == base {
		return false
	}

	return slices.Contains(statusClassRefinements[base], candidate)
}
