package lightfield

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Every error body is wrapped in an `error` object.
// https://docs.lightfield.app/using-the-api/errors
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"error"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	Err struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
		Param   string `json:"param,omitempty"`
	} `json:"error"`
}

func (r ResponseError) CombineErr(base error) error {
	details := make([]string, 0, 3) // nolint:mnd

	for _, part := range []string{r.Err.Type, r.Err.Code, r.Err.Param} {
		if part != "" {
			details = append(details, part)
		}
	}

	if r.Err.Message == "" && len(details) == 0 {
		return base
	}

	if len(details) == 0 {
		return fmt.Errorf("%w: %s", base, r.Err.Message)
	}

	return fmt.Errorf("%w: %s (%s)", base, r.Err.Message, strings.Join(details, ", "))
}
