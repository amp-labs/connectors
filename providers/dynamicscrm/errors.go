package dynamicscrm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	apiError := &CRMResponseError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return fmt.Errorf("interpretJSONError: %w %w", interpreter.ErrUnmarshal, err)
	}

	return createError(interpreter.DefaultStatusCodeMappingToErr(res, body), apiError)
}

type CRMResponseError struct {
	Err CRMError `json:"error"`
}

type CRMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	*EnhancedCRMError
}

type EnhancedCRMError struct {
	HelpLink     string `json:"@Microsoft.PowerApps.CDS.HelpLink"`           // nolint:tagliatelle
	TraceText    string `json:"@Microsoft.PowerApps.CDS.TraceText"`          // nolint:tagliatelle
	InnerMessage string `json:"@Microsoft.PowerApps.CDS.InnerError.Message"` // nolint:tagliatelle
}

func createError(base error, response *CRMResponseError) error {
	if len(response.Err.Message) > 0 {
		return fmt.Errorf("%w: %s", base, response.Err.Message)
	}

	return base
}
