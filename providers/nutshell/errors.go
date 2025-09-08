package nutshell

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"message", "code"},
			Template: func() interpreter.ErrorDescriptor { return &ErrorMessage{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ErrorDetails{} },
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusMethodNotAllowed:     common.ErrBadRequest,
	http.StatusUnsupportedMediaType: common.ErrBadRequest,
	http.StatusConflict:             common.ErrBadRequest,
}

var textErrorStatusCodeHandler = interpreter.DirectFaultyResponder{ // nolint:gochecknoglobals
	Callback: func(res *http.Response, body []byte) error {
		return fmt.Errorf("%w: %v",
			interpreter.StatusCodeMapper{Registry: statusCodeMapping}.MatchStatusCodeError(res, body),
			string(body),
		)
	},
}

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e ErrorMessage) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, e.Message)
}

type ErrorDetails struct {
	Meta     any             `json:"meta"`
	Links    any             `json:"links"`
	Accounts []any           `json:"accounts"`
	Errors   []errorResponse `json:"errors"`
}

type errorResponse struct {
	Id     any    `json:"id"`
	Type   string `json:"type"`
	Code   any    `json:"code"`
	Detail string `json:"detail"`
	Href   any    `json:"href"`
	Status string `json:"status"`
	Title  string `json:"title"`
	Field  any    `json:"field"`
}

func (d ErrorDetails) CombineErr(base error) error {
	reasons := make([]string, len(d.Errors))
	for i, item := range d.Errors {
		reasons[i] = fmt.Sprintf("%v [%v]", item.Title, item.Detail)
	}

	return fmt.Errorf("%w: %v", base, strings.Join(reasons, ","))
}

func (c *Connector) interpretHTMLError(res *http.Response, body []byte) error {
	base := interpreter.DefaultStatusCodeMappingToErr(res, body)

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		// ignore HTML that cannot be understood
		return base
	}

	message := document.Find("title").Text()

	return fmt.Errorf("%w: %v", base, message)
}
