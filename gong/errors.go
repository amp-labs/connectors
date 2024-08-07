package gong

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	formats := interpreter.NewFormatSwitch(
		[]interpreter.FormatTemplate{
			{
				MustKeys: nil,
				Template: &ResponseError{},
			},
		}...,
	)

	schema := formats.ParseJSON(body)

	return schema.CombineErr(interpreter.DefaultStatusCodeMappingToErr(res, body))
}

type ResponseError struct {
	RequestId string   `json:"requestId"`
	Errors    []string `json:"errors"`
}

func (r ResponseError) CombineErr(base error) error {
	if len(r.Errors) == 0 {
		return base
	}

	return fmt.Errorf("%w: %v", base, strings.Join(r.Errors, ","))
}
