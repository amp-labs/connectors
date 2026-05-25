package mail

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/providers/google/internal/core"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ErrorResponse{} },
		},
	}...,
)

// jsonResponder mirrors what was previously passed inline to ErrorHandler.JSON; it is
// now reused as the non-rate-limit fallback in interpretJSONError.
var jsonResponder = interpreter.NewFaultyResponder(errorFormats, nil) //nolint:gochecknoglobals

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  []struct {
		Message string `json:"message"`
		Domain  string `json:"domain"`
		Reason  string `json:"reason"`
	} `json:"errors"`
	Status string `json:"status"`
}

func (e ErrorResponse) CombineErr(base error) error {
	messages := make([]string, len(e.Error.Errors))
	for index, obj := range e.Error.Errors {
		messages[index] = obj.Message
	}

	if len(messages) == 0 {
		return fmt.Errorf("%w: %v", base, e.Error.Message)
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ","))
}

// interpretJSONError routes Gmail JSON error bodies through the shared Google
// rate-limit detector. Gmail returns 403 with reason RATE_LIMIT_EXCEEDED for
// per-user quota violations; without this hook the response maps to ErrForbidden
// (unrecoverable) instead of ErrLimitExceeded (retryable).
func (a *Adapter) interpretJSONError(res *http.Response, body []byte) error {
	return core.InterpretJSONError(res, body, jsonResponder.HandleErrorResponse)
}

func (a *Adapter) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	secondParagraph := document.Find("p").Eq(1)
	message := strings.TrimSpace(secondParagraph.Text())

	if message == "" {
		// Just use the generic title.
		title := document.Find("title").Text()

		return fmt.Errorf("%w: %v", base, title)
	}

	return fmt.Errorf("%w: %v", base, message)
}
