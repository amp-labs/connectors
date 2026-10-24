package circleback

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
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

// handleError interprets Circleback's error envelope and keeps the response's status and headers
// (common.HTTPError), so a caller can tell a missing transcript from other failures. Its
// rate-limit headers also pace the requests that follow.
func (c *Connector) handleError(response *http.Response, body []byte) error {
	c.pacer.observe(response.Header)

	err := interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle(response, body)

	return common.NewHTTPError(response.StatusCode, body, common.GetResponseHeaders(response), err)
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
