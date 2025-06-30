package facebook

import (
	"fmt"

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
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FbtraceId string `json:"fbtrace_id"`
	} `json:"error"`
}

func (r ResponseError) CombineErr(base error) error {
	// Error field is the safest to return, though not very useful.
	return fmt.Errorf("%w: %v", base, r.Error.Message)
}
