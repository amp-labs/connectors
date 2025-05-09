package capsule

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
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

type ResponseMessageError struct {
	Message string `json:"message"`
	Errors  []struct {
		Message  string `json:"message"`
		Resource string `json:"resource"`
		Field    string `json:"field"`
	} `json:"errors"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if len(r.Message) == 0 {
		return base
	}

	messages := make([]string, len(r.Errors))
	for index, subErr := range r.Errors {
		messages[index] = subErr.Message
	}

	var description string
	if len(messages) != 0 {
		description = fmt.Sprintf(" (%v)", strings.Join(messages, ","))
	}

	return fmt.Errorf("%w: %v%v", base, r.Message, description)
}

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}
