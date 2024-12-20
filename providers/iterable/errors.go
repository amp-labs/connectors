package iterable

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
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseError struct {
	Message string `json:"msg"`
	Code    string `json:"code"`
	Params  any    `json:"params"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Message)
}

func (c *Connector) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	// Several HTML errors have different formats, given a small sample -- it is safe to use
	// title as primary message.
	// For examples refer to the `test` directory which has `.html` samples used in unit testing.
	message := document.Find("title").Text()
	h2 := document.Find("h2").Text()
	detail := document.Find("#detail").Text()

	// Detail element has better message than Header2.
	errReason := h2
	if len(detail) != 0 {
		errReason = detail
	}

	if len(errReason) != 0 {
		message += ": " + errReason
	}

	message = cleanMessage(message)

	return fmt.Errorf("%w: %v", base, message)
}

func cleanMessage(message string) string {
	// Remove new lines and double spaces.
	r := strings.NewReplacer("\n", "", "  ", "")

	size := 0
	for size != len(message) {
		// Replace until nothing changes.
		size = len(message)
		message = r.Replace(message)
	}

	return message
}
