package sybill

import (
	"encoding/json"
	"fmt"

	"github.com/amp-labs/connectors/common/interpreter"
)

// GET endpoints return `{"detail": "<message>"}`; validation errors (422) return
// `{"detail": [{"loc": [...], "msg": "...", "type": "..."}]}`.
// https://api.sybill.ai/docs/endpoints.html#common-error-responses
var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	interpreter.FormatTemplate{
		MustKeys: []string{"detail"},
		Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
	},
)

type ResponseError struct {
	Detail json.RawMessage `json:"detail"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Detail) == 0 {
		return base
	}

	var message string
	if err := json.Unmarshal(r.Detail, &message); err == nil {
		return fmt.Errorf("%w: %s", base, message)
	}

	return fmt.Errorf("%w: %s", base, string(r.Detail))
}
