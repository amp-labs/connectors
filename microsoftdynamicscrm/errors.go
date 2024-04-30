package microsoftdynamicscrm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error { //nolint:cyclop
	apiError := &CRMResponseError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return fmt.Errorf("interpretJSONError: %w %w", interpreter.ErrUnmarshal, err)
	}

	switch res.StatusCode {
	case http.StatusBadRequest:
		return createError(common.ErrBadRequest, apiError)
	case http.StatusUnauthorized:
		return createError(common.ErrAccessToken, apiError)
	case http.StatusForbidden:
		return createError(common.ErrForbidden, apiError)
	case http.StatusNotFound:
		return createError(common.ErrBadRequest, apiError) // TODO more specific error
	case http.StatusMethodNotAllowed:
		return createError(common.ErrBadRequest, apiError) // TODO more specific error
	case http.StatusPreconditionFailed:
		return createError(common.ErrBadRequest, apiError) // TODO more specific error
	case http.StatusRequestEntityTooLarge:
		return createError(common.ErrBadRequest, apiError) // TODO more specific error
	case http.StatusTooManyRequests:
		return createError(common.ErrLimitExceeded, apiError)
	case http.StatusNotImplemented:
		return createError(common.ErrNotImplemented, apiError)
	case http.StatusServiceUnavailable:
		return createError(common.ErrServer, apiError)
	default:
		return common.InterpretError(res, body)
	}
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
