package loxo

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common/interpreter"
)

func (c *Connector) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	// Message is located under the <pre></pre> tag
	message := document.Find("head > title").Text()

	if message == "" {
		return base // nothing meaningful found, return base error
	}

	return fmt.Errorf("%w: %v", base, message)
}
