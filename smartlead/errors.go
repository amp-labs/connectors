package smartlead

import (
	"bytes"
	"fmt"
	"net/http"

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

type ResponseError struct{}

func (r ResponseError) CombineErr(base error) error {
	return base
}

func (c *Connector) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	// Message is located under the <pre></pre> tag
	message := document.Find("pre").Text()

	return fmt.Errorf("%w: %v", base, message)
}
