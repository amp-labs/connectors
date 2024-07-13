package zendesksupport

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

	var schema common.ErrorDescriptor

	if _, ok := payload["description"]; ok {
		apiError := &DescriptiveResponseError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError ListError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	} else {
		apiError := &MessageResponseError{}
		if err := json.Unmarshal(body, &apiError); err != nil {
			return fmt.Errorf("interpretJSONError SingleError: %w %w", interpreter.ErrUnmarshal, err)
		}

		schema = apiError
	}

	return schema.CombineErr(interpreter.DefaultStatusCodeMappingToErr(res, body))
}

type DescriptiveResponseError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

type MessageResponseError struct {
	Error struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	} `json:"error"`
}

func (r DescriptiveResponseError) CombineErr(base error) error {
	if len(r.Error)+len(r.Description) == 0 {
		return base
	}

	return fmt.Errorf("%w: [%v]%v", base, r.Error, r.Description)
}

func (r MessageResponseError) CombineErr(base error) error {
	if len(r.Error.Title)+len(r.Error.Message) == 0 {
		return base
	}

	return fmt.Errorf("%w: [%v]%v", base, r.Error.Title, r.Error.Message)
}
