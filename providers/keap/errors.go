package keap

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
	Detail string `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Detail)
}

func (c *Connector) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	// Several HTML errors have different formats, given a small sample -- it is safe to use
	// title as primary message; h2 may hold better description.
	// For examples refer to the `test` directory which has `.html` samples used in unit testing.
	message := document.Find("title").Text()
	h2 := document.Find("h2").Text()

	if len(h2) != 0 {
		message += ": " + h2
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
