package twenty

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

// Every error seen live is a JSON body with the HTTP status repeated in `statusCode`, an `error` code and a
// `messages` list (the OpenAPI document Twenty serves also shows a singular `message`):
//
//	{"statusCode":400,"error":"BadRequestException","messages":["object 'foobars' not found. eg: companies"]}
//	{"statusCode":429,"error":"RATE_LIMITED","messages":["Rate limit exceeded for apiKey: 100 requests per 60s."], ...}
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"statusCode"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	StatusCode int      `json:"statusCode"`
	Code       string   `json:"error"`
	Messages   []string `json:"messages"`
	Message    string   `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	messages := r.Messages
	if r.Message != "" {
		messages = append(messages, r.Message)
	}

	if len(messages) == 0 {
		return base
	}

	return fmt.Errorf("%w: %w", base, errors.New(strings.Join(messages, "; "))) // nolint:err113
}

// errorHandler maps a status to the shared sentinels (interpreter.DefaultStatusCodeMappingToErr: 400 and 404
// stop the request, 429 is ErrLimitExceeded, 5xx is ErrServer), adds Twenty's messages, and wraps the result
// with common.NewHTTPError so the status, body and headers (Retry-After, X-RateLimit-*) reach the server.
// common.InterpretError alone would class 404 as retryable: Twenty answers 404 for a record id that does not
// exist, which no retry can fix. Responses that are not JSON (a mistyped domain, a proxy in front of the
// instance answering in HTML) take the same mapping, with their body kept as the message.
// 408 is the one status taken out of the shared map, which stops the request on it: Twenty does not document
// a 408, so one comes from a proxy or load balancer in front of a self-hosted instance timing out, and a
// later attempt can succeed.
func errorHandler(res *http.Response, body []byte) error {
	responder := interpreter.NewFaultyResponder(errorFormats, map[int]error{
		http.StatusRequestTimeout: common.ErrRetryable,
	})
	handler := interpreter.DirectFaultyResponder{
		Callback: func(res *http.Response, body []byte) error {
			return common.NewHTTPError(res.StatusCode, body, common.GetResponseHeaders(res),
				responder.HandleErrorResponse(res, body))
		},
	}

	return interpreter.ErrorHandler{
		JSON:     handler,
		Fallback: handler,
	}.Handle(res, body)
}
