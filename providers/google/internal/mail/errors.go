package mail

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
			Template: func() interpreter.ErrorDescriptor { return &ErrorResponse{} },
		},
	}...,
)

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
