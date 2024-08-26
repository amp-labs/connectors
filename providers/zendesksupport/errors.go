package zendesksupport

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"description"},
			Template: func() interpreter.ErrorDescriptor { return &DescriptiveResponseError{} },
		}, {
			MustKeys: []string{"status"},
			Template: func() interpreter.ErrorDescriptor { return &StatusResponseError{} },
		}, {
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &MessageResponseError{} },
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusInternalServerError: common.ErrServer,
}

type DescriptiveResponseError struct {
	descrDetailsError
	Details map[string][]descrDetailsError `json:"details"`
}

type descrDetailsError struct {
	ErrorStr    string `json:"error"`
	Description string `json:"description"`
}

func (d descrDetailsError) Error() string {
	return fmt.Sprintf("[%v]%v", d.ErrorStr, d.Description)
}

type StatusResponseError struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

type MessageResponseError struct {
	Error struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r DescriptiveResponseError) CombineErr(base error) error {
	if len(r.ErrorStr)+len(r.Description) == 0 {
		return base
	}

	details := []error{
		r.descrDetailsError,
	}

	for _, list := range r.Details {
		for _, err := range list {
			details = append(details, err)
		}
	}

	return fmt.Errorf("%w: %w", base, errors.Join(details...))
}

func (r StatusResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Error)
}

func (r MessageResponseError) CombineErr(base error) error {
	if len(r.Error.Title)+len(r.Error.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: [%v]%v", base, r.Error.Title, r.Error.Message)
}
