package calendar

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ErrorDetails{} },
		},
	}...,
)

// ErrorDetails
// nolint:tagliatelle
type ErrorDetails struct {
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Errors  []errorItem `json:"errors"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

type errorItem struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (d ErrorDetails) CombineErr(base error) error {
	reasons := make([]string, len(d.Error.Errors))
	for i, item := range d.Error.Errors {
		reasons[i] = item.Message
	}

	message := strings.Join(reasons, ",")
	if len(message) == 0 {
		message = d.Error.Message
	}

	return fmt.Errorf("%w: %v", base, message)
}

// Typical HTML google error message has Title with paragraph describing the problem.
// For more format details check unit tests.
func (a *Adapter) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	secondParagraph := document.Find("p").Eq(1)
	// Remove the <ins> part to drop "That's all we know."
	secondParagraph.Find("ins").Remove()
	message := strings.TrimSpace(secondParagraph.Text())

	if message == "" {
		// Just use the generic title.
		title := document.Find("title").Text()

		return fmt.Errorf("%w: %v", base, title)
	}

	return fmt.Errorf("%w: %v", base, message)
}
