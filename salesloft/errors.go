package salesloft

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	apiError := &ResponseError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return fmt.Errorf("interpretJSONError: %w %w", interpreter.ErrUnmarshal, err)
	}

	return createError(interpreter.DefaultStatusCodeMappingToErr(res, body), apiError)
}

type ResponseError struct {
	Status int64  `json:"status"`
	Err    string `json:"error"`
}

func createError(base error, response *ResponseError) error {
	return fmt.Errorf("%w: %s", base, response.Err)
}
