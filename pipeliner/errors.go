package pipeliner

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
	var schema common.ErrorDescriptor

	if _, ok := payload["code"]; ok {
		apiError := &ResponseWithCodeError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError ResponseWithCodeError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	} else {
		// default to simple response
		apiError := &ResponseSimpleError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError ResponseSimpleError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	return schema.CombineErr(statusCodeMapping(res, body))
}

func statusCodeMapping(res *http.Response, body []byte) error {
	if res.StatusCode == http.StatusUnprocessableEntity {
		return common.ErrBadRequest
	}

	return interpreter.DefaultStatusCodeMappingToErr(res, body)
}

// ResponseSimpleError occurs for Read method, invalid URL.
type ResponseSimpleError struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	ErrorCode any    `json:"errorcode"`
	Traceback any    `json:"traceback"`
}

func (r ResponseSimpleError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}

// ResponseWithCodeError extended error format, happens during invalid Create/Update.
// nolint:tagliatelle
type ResponseWithCodeError struct {
	Code         int    `json:"code"`
	Name         string `json:"name"`
	Message      string `json:"message"`
	EntityId     string `json:"entity_id"`
	EntityName   string `json:"entity_name"`
	EntityErrors []struct {
		Code    int    `json:"code"`
		Name    string `json:"name"`
		Message string `json:"message"`
	} `json:"entity_errors"`
	FieldErrors []struct {
		FieldId   string `json:"field_id"`
		FieldName string `json:"field_name"`
		Name      string `json:"name"`
		Code      int    `json:"code"`
		Errors    []struct {
			Code      int    `json:"code"`
			Name      string `json:"name"`
			Message   string `json:"message"`
			FieldId   string `json:"field_id"`
			FieldName string `json:"field_name"`
		} `json:"errors"`
	} `json:"field_errors"`
	StepChecklistErrors []interface{} `json:"step_checklist_errors"`
	EntityIndex         interface{}   `json:"entity_index"`
	HTTPStatus          int           `json:"http_status"`
	Success             bool          `json:"success"`
}

func (r ResponseWithCodeError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}
