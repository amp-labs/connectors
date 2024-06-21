package gong

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	var payload ResponseError
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("interpretJSONError general: %w %w", interpreter.ErrUnmarshal, err)
	}

	return payload.CombineErr(statusCodeMapping(res, body))
}

func statusCodeMapping(res *http.Response, body []byte) error {
	switch res.StatusCode { // nolint:gocritic
	default:
		return interpreter.DefaultStatusCodeMappingToErr(res, body)
	}
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
