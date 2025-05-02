package fireflies

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Implement error abstraction layers to streamline provider error handling.
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

var (
	ErrMeetingLinkRequired           = errors.New("meeting link field is required")
	ErrUpdateMeetingLinkNotSupported = errors.New("updation of meeting link not supported")
	ErrStartTimeRequired             = errors.New("start time field is required for create bite")
	ErrEndTimeRequired               = errors.New("end time field is required for create bite")
	ErrUpdateBiteNotSupported        = errors.New("updation of bite is not supported")
	ErrRoleRequired                  = errors.New("role field is required")
	ErrUpdateRoleNotSupported        = errors.New("updating the role is not supported")
	ErrUpdateAudioSupported          = errors.New("updating the audio is not supported")
	ErrTitleRequired                 = errors.New("title field is required")
	ErrCreateMeetingSupported        = errors.New("creating the meeting is not supported")
	ErrInvalidResponseFormat         = errors.New("invalid input format")
	ErrURLIsRequired                 = errors.New("url field is required")
)

// ResponseError represents an error response from the fireflies API.
type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	Friendly   bool   `json:"friendly,omitempty"`
	Locations  any    `json:"locations,omitempty"`
	Extensions any    `json:"extensions,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for i, obj := range r.Errors {
		messages[i] = obj.Message
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}
