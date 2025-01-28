package google

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ErrorDetails{} },
		},
	}...,
)

// nolint:tagliatelle
type ErrorDetails struct {
	Error errorResponse `json:"error"`
}

type errorResponse struct {
	Errors  []errorItem `json:"errors"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

type errorItem struct {
	Domain  string `json:"domain"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (d ErrorDetails) CombineErr(base error) error {
	reasons := make([]string, len(d.Error.Errors))
	for i, item := range d.Error.Errors {
		reasons[i] = item.Message
	}

	message := strings.Join(reasons, ",")
	if len(message) == 0 {
		message = d.Error.Message
	}

	return fmt.Errorf("%w: %v", base, message)
}
