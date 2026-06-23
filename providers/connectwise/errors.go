package connectwise

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

// ResponseError
// nolint:tagliatelle
type ResponseError struct {
	Message string `json:"message"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (r ResponseError) CombineErr(base error) error {
	messages := make([]string, 1+len(r.Errors))

	messages[0] = r.Message
	for index, object := range r.Errors {
		messages[index+1] = object.Message
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ": "))
}
