package core

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func createError(baseErr error, sfErr jsonError, res *http.Response) error {
	if len(sfErr.Message) > 0 {
		return fmt.Errorf("%w: %s (HTTP status %d)", baseErr, sfErr.Message, res.StatusCode)
	}

	return baseErr
}

func interpretJSONError(res *http.Response, body []byte) error { // nolint:cyclop
	var errs []jsonError
	if err := json.Unmarshal(body, &errs); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	for _, sfErr := range errs {
		switch sfErr.ErrorCode {
		case "INVALID_SESSION_ID":
			return createError(common.ErrInvalidSessionId, sfErr, res)
		case "INSUFFICIENT_ACCESS_OR_READONLY":
			return createError(common.ErrForbidden, sfErr, res)
		case "API_DISABLED_FOR_ORG":
			return createError(common.ErrApiDisabled, sfErr, res)
		case "UNABLE_TO_LOCK_ROW":
			return createError(common.ErrUnableToLockRow, sfErr, res)
		case "INVALID_GRANT":
			return createError(common.ErrInvalidGrant, sfErr, res)
		case "REQUEST_LIMIT_EXCEEDED":
			return createError(common.ErrLimitExceeded, sfErr, res)
		case "INVALID_TYPE":
			fallthrough
		case "INVALID_FIELD_FOR_INSERT_UPDATE":
			fallthrough
		case "MALFORMED_QUERY":
			fallthrough
		case "FIELD_INTEGRITY_EXCEPTION":
			fallthrough
		case "INVALID_FIELD":
			return createError(common.ErrBadRequest, sfErr, res)
		case "INVALID_QUERY_LOCATOR":
			return createError(common.ErrCursorGone, sfErr, res)
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
	}, res)
}
