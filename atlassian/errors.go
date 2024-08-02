package atlassian

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

	if _, ok := payload["status"]; ok {
		apiError := &ResponseStatusError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError SingleError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	} else {
		apiError := &ResponseMessagesError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError MessagesError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	// enhance status code error with response payload
	return schema.CombineErr(interpreter.DefaultStatusCodeMappingToErr(res, body))
}

type ResponseMessagesError struct {
	ErrorMessages   []string `json:"errorMessages"`
	WarningMessages []string `json:"warningMessages"`
}

func (r ResponseMessagesError) CombineErr(base error) error {
	if len(r.ErrorMessages) == 0 {
		return base
	}

	message := strings.Join(r.ErrorMessages, ",")

	return fmt.Errorf("%w: %v", base, message)
}

type ResponseStatusError struct {
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

func (r ResponseStatusError) CombineErr(base error) error {
	if len(r.Error) == 0 {
		return base
	}

	if len(r.Message) == 0 {
		return fmt.Errorf("%w: %v", base, r.Error)
	}

	return fmt.Errorf("%w: %v - %v", base, r.Error, r.Message)
}
