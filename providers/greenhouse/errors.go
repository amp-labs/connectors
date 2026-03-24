package greenhouse

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"message"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

type ResponseMessageError struct {
	Message string           `json:"message"`
	Errors  []ResponseDetail `json:"errors"`
}

type ResponseDetail struct {
	Message string `json:"message"`
	Field   string `json:"field"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Message) == 0 {
		return base
	}

	messages := make([]string, 0, len(r.Errors))
	for _, subErr := range r.Errors {
		messages = append(messages, subErr.Message)
	}

	if len(messages) == 0 {
		return fmt.Errorf("%w: %v", base, r.Message)
	}

	return fmt.Errorf("%w: %v (%v)", base, r.Message, strings.Join(messages, ", "))
}

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}
