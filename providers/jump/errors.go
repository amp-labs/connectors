package jump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// https://my.jumpapp.com/enterprise/documentation (Errors)
var errorFormats = interpreter.NewFormatSwitch( //nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			// Authentication and rate limiting happen before GraphQL executes.
			MustKeys: []string{"error"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseAuthError{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

// graphqlErrorResponder formats GraphQL error payloads into typed errors,
// reusing the same error schema as the standard (non-2xx) error handler.
var graphqlErrorResponder = interpreter.NewFaultyResponder(errorFormats, nil) //nolint:gochecknoglobals

// ResponseAuthError is returned with a non-2xx status before GraphQL executes.
//
//	{
//		"error": {
//		  "code": "RATE_LIMITED",
//		  "message": "Rate limit exceeded",
//		  "details": {
//			"retry_after": 23
//		  }
//		}
//	  }
type ResponseAuthError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r ResponseAuthError) CombineErr(base error) error {
	if r.Error.Message == "" {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Error.Message)
}

// ResponseError is the standard GraphQL errors array.
//
//	{
//		"errors": [{
//		  "message": "Meeting not found",
//		  "extensions": {
//			"code": "MEETING_NOT_FOUND",
//			"details": {
//			  "id": "mtg_123"
//			}
//		  }
//		}]
//	  }
type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message    string `json:"message,omitempty"`
	Extensions struct {
		Code string `json:"code,omitempty"`
	} `json:"extensions"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		messages[i] = obj.Message

		if obj.Extensions.Code == "INVALID_PAGINATION" {
			base = fmt.Errorf("%w: %w", base, common.ErrInvalidPaginationCursor)
		}
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}

// graphqlResponse captures just enough of a GraphQL payload to decide
// whether a 200 response is actually a failure.
type graphqlResponse struct {
	Errors []ErrorDetails             `json:"errors"`
	Data   map[string]json.RawMessage `json:"data"`
}

// interpretGraphQLError returns a typed error when a 2xx GraphQL response is actually a failure.
// Jump executes GraphQL after authentication and reports execution errors, such as complexity
// limits, invalid cursors or unknown records, with a 200 status and an "errors" array.
// The response is a failure when the queried object is missing or null.
func interpretGraphQLError(resp *common.JSONHTTPResponse, objectName string) error {
	node, ok := resp.Body()
	if !ok {
		return nil
	}

	body := node.Source()

	var payload graphqlResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil //nolint:nilerr
	}

	data := payload.Data[objectName]
	hasData := len(data) > 0 && !bytes.Equal(data, []byte("null"))

	if len(payload.Errors) == 0 || hasData {
		return nil
	}

	synthetic := &http.Response{
		StatusCode: graphqlStatusCode(payload.Errors[0]),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	return graphqlErrorResponder.HandleErrorResponse(synthetic, body)
}

// graphqlStatusCode maps Jump GraphQL errors to HTTP status codes understood by the error interpreter.
// Errors raised while validating the query, such as an exceeded complexity, carry no code.
func graphqlStatusCode(details ErrorDetails) int {
	code := details.Extensions.Code

	switch {
	case strings.HasSuffix(code, "_NOT_FOUND"):
		return http.StatusNotFound
	case code == "INTERNAL_ERROR":
		return http.StatusInternalServerError
	case code == "RATE_LIMITED":
		return http.StatusTooManyRequests
	case code == "ACCESS_DENIED", details.Message == "Unauthorized":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}
