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
	Code   string `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
	Id     string `json:"id,omitempty"`
	Source Source `json:"source"`
	Status string `json:"status,omitempty"`
	Title  string `json:"title,omitempty"`
}

type Source struct {
	Cookie    string `json:"cookie,omitempty"`
	Header    string `json:"header,omitempty"`
	Parameter string `json:"parameter,omitempty"`
	Path      string `json:"path,omitempty"`
	Pointer   string `json:"pointer,omitempty"`
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
