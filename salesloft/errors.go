package salesloft

import (
	"encoding/json"
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

	// now we can choose which error response Schema we expect
	var schema errorDescriptor

	if _, ok := payload["errors"]; ok {
		apiError := &ResponseListError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError ListError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
		if res.StatusCode == http.StatusUnprocessableEntity {
			return schema.combineErr(common.ErrBadRequest)
		}
	} else {
		// default to simple response
		apiError := &ResponseSingleError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError SingleError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	// enhance status code error with response payload
	return schema.combineErr(interpreter.DefaultStatusCodeMappingToErr(res, body))
}

type errorDescriptor interface {
	combineErr(base error) error
}

type ResponseListError struct {
	Errors map[string]any `json:"errors"`
}

func (r ResponseListError) combineErr(base error) error {
	var message string

	data, err := json.Marshal(r.Errors)
	if err != nil {
		message = "failed parsing error response"
	} else {
		message = string(data)
	}

	return fmt.Errorf("%w: %s", base, message)
}

type ResponseSingleError struct {
	Status int64  `json:"status"`
	Err    string `json:"error"`
}

func (r ResponseSingleError) combineErr(base error) error {
	return fmt.Errorf("%w: %s", base, r.Err)
}
