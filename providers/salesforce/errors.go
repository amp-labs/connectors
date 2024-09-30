package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/xquery"
)

var ErrCannotReadMetadata = errors.New("cannot read object metadata, it is possible you don't have the correct permissions set") // nolint:lll

type jsonError struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
}

func createError(baseErr error, sfErr jsonError) error {
	if len(sfErr.Message) > 0 {
		return fmt.Errorf("%w: %s", baseErr, sfErr.Message)
	}

	return baseErr
}

func (c *Connector) interpretJSONError(res *http.Response, body []byte) error { // nolint:cyclop
	var errs []jsonError
	if err := json.Unmarshal(body, &errs); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	for _, sfErr := range errs {
		switch sfErr.ErrorCode {
		case "INVALID_SESSION_ID":
			return createError(common.ErrInvalidSessionId, sfErr)
		case "INSUFFICIENT_ACCESS_OR_READONLY":
			return createError(common.ErrForbidden, sfErr)
		case "API_DISABLED_FOR_ORG":
			return createError(common.ErrApiDisabled, sfErr)
		case "UNABLE_TO_LOCK_ROW":
			return createError(common.ErrUnableToLockRow, sfErr)
		case "INVALID_GRANT":
			return createError(common.ErrInvalidGrant, sfErr)
		case "REQUEST_LIMIT_EXCEEDED":
			return createError(common.ErrLimitExceeded, sfErr)
		case "INVALID_TYPE":
			fallthrough
		case "INVALID_FIELD_FOR_INSERT_UPDATE":
			fallthrough
		case "MALFORMED_QUERY":
			fallthrough
		case "FIELD_INTEGRITY_EXCEPTION":
			fallthrough
		case "INVALID_FIELD":
			return createError(common.ErrBadRequest, sfErr)
		default:
			continue
		}
	}

	// No known errors, just do the normal error handling logic
	return common.InterpretError(res, body)
}

func (c *Connector) interpretXMLError(res *http.Response, body []byte) error {
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
	})
}
