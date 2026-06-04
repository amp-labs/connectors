package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/xquery"
)

func NewErrorHandler() *interpreter.ErrorHandler {
	return &interpreter.ErrorHandler{
		JSON: &interpreter.DirectFaultyResponder{Callback: interpretJSONError},
		XML:  &interpreter.DirectFaultyResponder{Callback: interpretXMLError},
	}
}

type jsonError struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

func createError(baseErr error, sfErr jsonError, res *http.Response, body []byte) error {
	var innerErr error
	if len(sfErr.Message) > 0 {
		innerErr = fmt.Errorf("%w: %s", baseErr, sfErr.Message)
	} else {
		innerErr = baseErr
	}

	return common.NewHTTPError(res.StatusCode, body, common.GetResponseHeaders(res), innerErr)
}

// noSuchColumnRe extracts the field and object names from a Salesforce
// "No such column" error message. Salesforce phrases the object differently
// depending on the API path: reads use "on entity 'Contact'" while writes
// use "on sobject of type Contact" (unquoted).
var noSuchColumnRe = regexp.MustCompile(
	`No such column '([^']*)' on (?:entity '([^']*)'|sobject of type (\w+))`)

// fieldNotFoundGuidance explains the two causes of a "No such column" error
// (incorrect field name, or missing field-level visibility) It is appended to
// every formatted field-not-found message.
//
//nolint:lll
const fieldNotFoundGuidance = " This usually means either the field name is incorrect (custom field names must end in '__c'), or the connected Salesforce user lacks field-level visibility for this field."

// formatFieldNotFoundMessage turns a Salesforce "No such column" error into a
// customer-facing message: it restates the problem in Salesforce admin
// vocabulary (field/object) and appends actionable guidance. When the field
// and object names cannot be extracted, it falls back to Salesforce's own
// sentence with its trailing noise stripped.
func formatFieldNotFoundMessage(msg string) string {
	if m := noSuchColumnRe.FindStringSubmatch(msg); m != nil {
		object := m[2]
		if object == "" {
			object = m[3]
		}

		return fmt.Sprintf("Field '%s' was not found or is not accessible on object '%s'.%s",
			m[1], object, fieldNotFoundGuidance)
	}

	if i := strings.Index(msg, "No such column"); i >= 0 {
		msg = msg[i:]
	}

	msg = strings.TrimSpace(msg)
	msg = strings.TrimSuffix(msg,
		" Please reference your WSDL or the describe call for the appropriate names.")
	msg = strings.TrimSuffix(strings.TrimSpace(msg),
		" If you are attempting to use a custom field, be sure to append the '__c' after the custom field name.")

	return strings.TrimSpace(msg) + fieldNotFoundGuidance
}

// fieldNotFoundError wraps common.ErrBadRequest for errors.Is matching while
// rendering only the supplied message. It is intentionally not returned via
// createError / common.NewHTTPError so the formatted, customer-facing message
// is not prefixed with "HTTP status N: " or "bad request: ". The trade-off is
// that the raw provider response body is not propagated for this error case.
type fieldNotFoundError struct {
	msg string
}

func (e *fieldNotFoundError) Error() string { return e.msg }

// Unwrap returns both sentinels so callers can match the specific
// ErrFieldNotFound case while existing errors.Is(err, ErrBadRequest) checks
// continue to hold.
func (e *fieldNotFoundError) Unwrap() []error {
	return []error{common.ErrFieldNotFound, common.ErrBadRequest}
}

func interpretJSONError(res *http.Response, body []byte) error { // nolint:cyclop
	var errs []jsonError
	if err := json.Unmarshal(body, &errs); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	for _, sfErr := range errs {
		switch sfErr.ErrorCode {
		case "INVALID_SESSION_ID":
			return createError(common.ErrInvalidSessionId, sfErr, res, body)
		case "INSUFFICIENT_ACCESS_OR_READONLY":
			return createError(common.ErrForbidden, sfErr, res, body)
		case "API_DISABLED_FOR_ORG":
			return createError(common.ErrApiDisabled, sfErr, res, body)
		case "UNABLE_TO_LOCK_ROW":
			return createError(common.ErrUnableToLockRow, sfErr, res, body)
		case "INVALID_GRANT":
			return createError(common.ErrInvalidGrant, sfErr, res, body)
		case "REQUEST_LIMIT_EXCEEDED":
			return createError(common.ErrLimitExceeded, sfErr, res, body)
		case "INVALID_TYPE":
			fallthrough
		case "INVALID_FIELD_FOR_INSERT_UPDATE":
			fallthrough
		case "MALFORMED_QUERY":
			fallthrough
		case "FIELD_INTEGRITY_EXCEPTION":
			fallthrough
		case "INVALID_FIELD":
			if strings.Contains(sfErr.Message, "No such column") {
				return &fieldNotFoundError{
					msg: formatFieldNotFoundMessage(sfErr.Message),
				}
			}

			return createError(common.ErrBadRequest, sfErr, res, body)
		case "INVALID_QUERY_LOCATOR":
			return createError(common.ErrCursorGone, sfErr, res, body)
		default:
			continue
		}
	}

	// No known errors, just do the normal error handling logic
	return common.InterpretError(res, body)
}

func interpretXMLError(res *http.Response, body []byte) error {
	xml, err := xquery.NewXML(body)
	if err != nil {
		// Response body cannot be understood in the form of valid XML structure.
		// Default error handling.
		return common.InterpretError(res, body)
	}

	code := xml.FindOne("//faultcode").Text()
	message := xml.FindOne("//faultstring").Text()

	var matchingErr error

	switch code {
	case "soapenv:Client":
		matchingErr = common.ErrBadRequest
	case "sf:INVALID_SESSION_ID":
		matchingErr = common.ErrAccessToken
	}

	if matchingErr == nil {
		return common.InterpretError(res, body)
	}

	return createError(matchingErr, jsonError{
		Message:   message,
		ErrorCode: code,
	}, res, body)
}
