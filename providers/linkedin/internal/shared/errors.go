package shared //nolint:revive

import (
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

var ErrorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

var ErrMetadataNotFound = errors.New("metadata not found")

type ResponseError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (r ResponseError) CombineErr(base error) error {
	// Error field is the safest to return, though not very useful.
	return fmt.Errorf("%w: %v", base, r.Message)
}
