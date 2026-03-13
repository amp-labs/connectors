package confluence

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"errors"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseCommonError{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

type ResponseCommonError struct {
	Errors []struct {
		Status int    `json:"status"`
		Code   string `json:"code"`
		Title  string `json:"title"`
		Detail any    `json:"detail"`
	} `json:"errors"`
}

func (r ResponseCommonError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for index, subErr := range r.Errors {
		messages[index] = subErr.Title
	}

	var descriptions string
	if len(messages) != 0 {
		descriptions = strings.Join(messages, ",")
	}

	return fmt.Errorf("%w: %v", base, descriptions)
}

type ResponseMessageError struct {
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Message)
}
