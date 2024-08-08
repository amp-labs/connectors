package pipeliner

import (
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: []string{"code"},
			Template: &ResponseWithCodeError{},
		}, {
			MustKeys: nil,
			Template: &ResponseSimpleError{},
		},
	}...,
)

var statusCodeMapping = map[int]error{ // nolint:gochecknoglobals
	http.StatusUnprocessableEntity: common.ErrBadRequest,
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
