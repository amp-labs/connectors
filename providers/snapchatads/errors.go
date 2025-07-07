package snapchatads

import (
	"errors"
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

var (
	ErrObjNotFound    = errors.New("object not found")
	DeleteResponseKey = "sub_request_error_reason"
)

type ResponseError struct {
	RequestStatus  string `json:"request_status"`  //nolint:tagliatelle
	RequestId      string `json:"request_id"`      //nolint:tagliatelle
	DebugMessage   string `json:"debug_message"`   //nolint:tagliatelle
	DisplayMessage string `json:"display_message"` //nolint:tagliatelle
	ErrorCode      string `json:"error_code"`      //nolint:tagliatelle
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: [%v] %v", base, r.ErrorCode, r.DisplayMessage)
}
