package docusign

import (
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"errorCode", "message"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseError struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v: %v", base, r.ErrorCode, r.Message)
}
