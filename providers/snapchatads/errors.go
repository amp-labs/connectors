package snapchatads

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
	RequestStatus  string `json:"request_status"`
	RequestId      string `json:"request_id"`
	DebugMessage   string `json:"debug_message"`
	DisplayMessage string `json:"display_message"`
	ErrorCode      string `json:"error_code"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: [%v] %v", base, r.ErrorCode, r.DisplayMessage)
}
