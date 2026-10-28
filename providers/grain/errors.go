package grain

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

// Every JSON failure carries an "error" code, with details in "error_description"
// (authentication) or "data" (request validation):
//
//	{"error": "invalid_token", "error_description": "The access token is invalid or expired."}
//	{"error": "bad_request", "data": "invalid_cursor"}
//	{"error": "not_found"}
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"error"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	Data             string `json:"data,omitempty"`
}

func (r ResponseError) CombineErr(base error) error {
	details := r.ErrorDescription
	if details == "" {
		details = strings.TrimSpace(r.Data)
	}

	if details == "" {
		return fmt.Errorf("%w: %s", base, r.Error)
	}

	return fmt.Errorf("%w: %s: %s", base, r.Error, details)
}
