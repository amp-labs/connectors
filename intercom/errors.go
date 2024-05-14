package intercom

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("interpretJSONError general: %w %w", interpreter.ErrUnmarshal, err)
	}

	// TODO handle error

	return interpreter.DefaultStatusCodeMappingToErr(res, body)
}
