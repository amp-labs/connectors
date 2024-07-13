package zendesksupport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("interpretJSONError general: %w %w", interpreter.ErrUnmarshal, err)
	}

	var schema common.ErrorDescriptor

	if _, ok := payload["description"]; schema == nil && ok {
		apiError := &DescriptiveResponseError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf(
				"interpretJSONError DescriptiveResponseError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	if _, ok := payload["status"]; schema == nil && ok {
		apiError := &StatusResponseError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError StatusResponseError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	// default format
	if schema == nil {
		apiError := &MessageResponseError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError MessageResponseError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	return schema.CombineErr(statusCodeMapping(res, body))
}

func statusCodeMapping(res *http.Response, body []byte) error {
	if res.StatusCode == http.StatusInternalServerError {
		return common.ErrServer
	}

	return interpreter.DefaultStatusCodeMappingToErr(res, body)
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
