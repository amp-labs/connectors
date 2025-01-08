package smartleadv2

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
			MustKeys: []string{"message"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		}, {
			MustKeys: []string{"error"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseBasicError{} },
		},
	}...,
)

type ResponseMessageError struct {
	Message string `json:"message"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Message) != 0 {
		return fmt.Errorf("%w: %s", base, r.Message)
	}

	return base
}

type ResponseBasicError struct {
	Error string `json:"error"`
}

func (r ResponseBasicError) CombineErr(base error) error {
	if len(r.Error) != 0 {
		return fmt.Errorf("%w: %s", base, r.Error)
	}

	return base
}

func interpretHTMLError(res *http.Response, body []byte) error {
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
