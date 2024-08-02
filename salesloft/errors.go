package salesloft

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	formats := interpreter.NewFormatSwitch(
		[]interpreter.FormatTemplate{
			{
				MustKeys: []string{"errors"},
				Template: &ResponseListError{},
			}, {
				MustKeys: nil,
				Template: &ResponseSingleError{},
			},
		}...,
	)

	schema, err := formats.ParseJSON(body)
	if err != nil {
		return err
	}

	return schema.CombineErr(statusCodeMapping(res, body))
}

func statusCodeMapping(res *http.Response, body []byte) error {
	if res.StatusCode == http.StatusUnprocessableEntity {
		return common.ErrBadRequest
	}

	return interpreter.DefaultStatusCodeMappingToErr(res, body)
}

type ResponseListError struct {
	Errors map[string]any `json:"errors"`
}

func (r ResponseListError) CombineErr(base error) error {
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

func (r ResponseSingleError) CombineErr(base error) error {
	return fmt.Errorf("%w: %s", base, r.Err)
}
