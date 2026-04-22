package gong

import (
	"fmt"
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

type ResponseError struct {
	RequestId string   `json:"requestId"`
	Errors    []string `json:"errors"`
}

// cursorErrorMessages contains known Gong error messages that indicate
// a pagination cursor has expired or is no longer valid.
var cursorErrorMessages = []string{ // nolint:gochecknoglobals
	"cursor has expired",
	"failed to verify cursor",
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	joined := strings.Join(r.Errors, ", ")

	for _, e := range r.Errors {
		lower := strings.ToLower(e)
		for _, msg := range cursorErrorMessages {
			if strings.Contains(lower, msg) {
				return fmt.Errorf("%w: %v", common.ErrCursorGone, joined)
			}
		}
	}

	return fmt.Errorf("%w: %v", base, joined)
}
