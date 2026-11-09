package auphan

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// Every failure shares one envelope, and Auphan may send it with HTTP 200:
// {"success": false, "error": {"code": "INVALID_PARAM", "message": "...", "details": {...}}}.
// https://api1.auphansoftware.com/resources/#authentication
var errorFormats = interpreter.NewFormatSwitch( //nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"error"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

var jsonErrorHandler = interpreter.ErrorHandler{ //nolint:gochecknoglobals
	JSON: interpreter.NewFaultyResponder(errorFormats, nil),
}.Handle

// interpretError handles non-2xx responses. Failures sent with HTTP 200 are caught by checkEnvelope.
func interpretError(response *http.Response, body []byte) error {
	// The gateway rejects an unknown token itself, with a plain-text "Forbidden" 403 (seen live),
	// not the documented 401 INVALID_TOKEN. Store-level denials are JSON, so they stay ErrForbidden.
	if response.StatusCode == http.StatusForbidden && !json.Valid(body) {
		return common.NewHTTPError(response.StatusCode, body, common.GetResponseHeaders(response),
			fmt.Errorf("%w: gateway rejected the token", common.ErrAccessToken))
	}

	return jsonErrorHandler(response, body)
}

type ResponseError struct {
	Success bool      `json:"success"`
	Error   *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	if r.Error == nil {
		return base
	}

	return fmt.Errorf("%w: %s (%s)", base, r.Error.Message, r.Error.Code)
}

// checkEnvelope turns {"success": false} in a 2xx response into an error.
func checkEnvelope(response *common.JSONHTTPResponse) error {
	envelope, err := common.UnmarshalJSON[ResponseError](response)
	if err != nil {
		return err
	}

	if envelope.Success {
		return nil
	}

	if envelope.Error == nil {
		return fmt.Errorf("%w: response has success=false and no error", common.ErrRequestFailed)
	}

	return envelope.CombineErr(errorFromCode(envelope.Error.Code))
}

// errorFromCode maps Auphan error codes to common errors.
// The codes come from the API reference and manual testing.
func errorFromCode(code string) error {
	switch code {
	case "MISSING_TOKEN", "INVALID_TOKEN":
		return common.ErrAccessToken
	case "USER_DISABLED", "STORE_ACCESS_DENIED", "MODULE_ACCESS_DENIED", "ACCESS_DENIED", "GROUP_NOT_ASSIGNABLE":
		return common.ErrForbidden
	case "NOT_FOUND":
		return fmt.Errorf("%w: %w", common.ErrBadRequest, common.ErrNotFound)
	case "INVALID_PARAM", "VALIDATION_ERROR", "INVALID_REQUEST", "UNKNOWN_FILTER":
		return common.ErrBadRequest
	default:
		return common.ErrRequestFailed
	}
}
