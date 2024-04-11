package msdsales

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
)

func (*Connector) interpretJSONError(res *http.Response, body []byte) error {
	apiError := &SalesResponseError{}
	if err := json.Unmarshal(body, &apiError); err != nil {
		return fmt.Errorf("interpretJSONError: %w %w", interpreter.ErrUnmarshal, err)
	}

	switch res.StatusCode {
	case http.StatusBadRequest:
		return errors.Join(common.ErrBadRequest, apiError)
	case http.StatusUnauthorized:
		return errors.Join(common.ErrAccessToken, apiError)
	case http.StatusForbidden:
		return errors.Join(common.ErrForbidden, apiError)
	case http.StatusNotFound:
		return errors.Join(common.ErrBadRequest, apiError) // FIXME more specific error
	case http.StatusMethodNotAllowed:
		return errors.Join(common.ErrBadRequest, apiError) // FIXME more specific error
	case http.StatusPreconditionFailed:
		return errors.Join(common.ErrBadRequest, apiError) // FIXME more specific error
	case http.StatusRequestEntityTooLarge:
		return errors.Join(common.ErrBadRequest, apiError) // FIXME more specific error
	case http.StatusTooManyRequests:
		return errors.Join(common.ErrLimitExceeded, apiError)
	case http.StatusNotImplemented:
		return errors.Join(common.ErrNotImplemented, apiError)
	case http.StatusServiceUnavailable:
		return errors.Join(common.ErrServer, apiError)
	default:
		return common.InterpretError(res, body)
	}
}

type SalesResponseError struct {
	Err SalesError `json:"error"`
}

type SalesError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// FIXME below fields are non empty only if request had header `Prefer: odata.include-annotations="*"`
	*EnhancedSalesError
}

type EnhancedSalesError struct {
	HelpLink     string `json:"@Microsoft.PowerApps.CDS.HelpLink"`
	TraceText    string `json:"@Microsoft.PowerApps.CDS.TraceText"`
	InnerMessage string `json:"@Microsoft.PowerApps.CDS.InnerError.Message"`
}

func (s SalesResponseError) Error() string {
	return s.Err.Message
}
