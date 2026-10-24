package circleback

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Failures share one envelope: {"error": "...", "code": "RATE_LIMITED", "issues": [...]}.
// https://circleback.ai/docs/api#rate-limits (ErrorResponse in the OpenAPI spec)
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"error"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

// handleError interprets Circleback's error envelope, and keeps the Retry-After of a 429 so a
// rate-limited request can be retried when the window refills.
// https://circleback.ai/docs/api#rate-limits
func handleError(response *http.Response, body []byte) error {
	err := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle(response, body)

	if response.StatusCode != http.StatusTooManyRequests {
		return err
	}

	seconds, convErr := strconv.Atoi(response.Header.Get("Retry-After"))
	if convErr != nil || seconds < 0 {
		return err
	}

	return &rateLimitError{err: err, retryAfter: time.Duration(seconds) * time.Second}
}

type rateLimitError struct {
	err        error
	retryAfter time.Duration
}

func (e *rateLimitError) Error() string {
	return fmt.Sprintf("%v (retry after %v)", e.err, e.retryAfter)
}

func (e *rateLimitError) Unwrap() error {
	return e.err
}

type ResponseError struct {
	Error  string  `json:"error"`
	Code   string  `json:"code,omitempty"`
	Issues []issue `json:"issues,omitempty"`
}

type issue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	message := r.Error

	if r.Code != "" {
		message = fmt.Sprintf("%s (%s)", message, r.Code)
	}

	if len(r.Issues) != 0 {
		details := make([]string, 0, len(r.Issues))

		for _, item := range r.Issues {
			details = append(details, fmt.Sprintf("%s: %s", item.Path, item.Message))
		}

		message = fmt.Sprintf("%s: %s", message, strings.Join(details, "; "))
	}

	if message == "" {
		return base
	}

	return fmt.Errorf("%w: %s", base, message)
}
