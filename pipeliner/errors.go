package pipeliner

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	var payload ResponseError
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("interpretJSONError general: %w %w", interpreter.ErrUnmarshal, err)
	}

	return payload.CombineErr(interpreter.DefaultStatusCodeMappingToErr(res, body))
}

type ResponseError struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	ErrorCode any    `json:"errorcode"`
	Traceback any    `json:"traceback"`
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: %v", base, r.Message)
}
