package klaviyo

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
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

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusConflict: common.ErrBadRequest,
}

type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Source any    `json:"source"`
	Meta   any    `json:"meta"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		messages[i] = obj.makeMessage()
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}

func (d ErrorDetails) makeMessage() string {
	if len(d.Detail) == 0 {
		return d.Title
	}

	if d.Title == d.Detail {
		return d.Title
	}

	title, _ := strings.CutSuffix(d.Title, ".") // remove dot at the end.

	return title + ": " + d.Detail
}
