package activecampaign

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
			MustKeys: []string{"errors"},
			Template: func() interpreter.ErrorDescriptor { return &ResponseListError{} },
		},
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseMessageError{} },
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
}

// ResponseListError is the JSON:API style error list returned by most endpoints.
// Example: {"errors":[{"title":"Contact Email Address is not valid.","detail":"","code":"email_invalid"}]}.
type ResponseListError struct {
	Errors []ErrorDetail `json:"errors"`
}

type ErrorDetail struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Code   string `json:"code"`
}

func (r ResponseListError) CombineErr(base error) error {
	messages := make([]string, 0, len(r.Errors))

	for _, descr := range r.Errors {
		message := descr.Title
		if descr.Detail != "" {
			message += ": " + descr.Detail
		}

		if message != "" {
			messages = append(messages, message)
		}
	}

	if len(messages) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, strings.Join(messages, ", "))
}

// ResponseMessageError covers auth failures.
// Example: {"message":"No Result found for Subscriber with id 1"}.
type ResponseMessageError struct {
	Message string `json:"message"`
}

func (r ResponseMessageError) CombineErr(base error) error {
	if r.Message == "" {
		return base
	}

	return fmt.Errorf("%w: %v", base, r.Message)
}
