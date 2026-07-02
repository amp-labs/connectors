package jump

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( //nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

type ResponseError struct {
	Errors []ErrorDetails `json:"errors"`
}

type ErrorDetails struct {
	Message string `json:"message,omitempty"`
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

//	{
//		"errors": [{
//		  "message": "Meeting not found",
//		  "extensions": {
//			"code": "MEETING_NOT_FOUND",
//			"details": {
//			  "id": "mtg_123"
//			}
//		  }
//		}]
//	  }
//
// Jump returns HTTP 200 even when the operation fails, so errors must be detected here rather
// than by the status-code based error handler.
func checkErrorInResponse(errorResp *ResponseError) error {
	if errorResp == nil || len(errorResp.Errors) == 0 {
		return nil
	}

	return errorResp.CombineErr(common.ErrBadRequest)
}
