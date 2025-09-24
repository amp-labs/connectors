package chorus

import (
	"fmt"
	"strings"

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
	Errors []Error `json:"errors"`
}

type Error struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
	Id     string `json:"id"`
	Source Source `json:"source"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

type Source struct {
	Cookie    string `json:"Cookie"`
	Header    string `json:"header"`
	Parameter string `json:"parameter"`
	Path      string `json:"path"`
	Pointer   string `json:"pointer"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	details := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		details[i] = obj.Detail
	}

	return fmt.Errorf("%w: %v", base, strings.Join(details, ", "))
}
