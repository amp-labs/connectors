package restlet

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

var ErrRestletError = errors.New("restlet error")

// checkResponseError inspects a RESTlet response envelope and returns an error
// if header.status is not SUCCESS.
func checkResponseError(data []byte) error {
	var resp restletResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal restlet response: %w", err)
	}

	if resp.Header.Status == "SUCCESS" {
		return nil
	}

	// Parse error details from body.
	var errBody restletErrorBody
	if err := json.Unmarshal(resp.Body, &errBody); err != nil {
		return fmt.Errorf("%w: status=%s (could not parse error body)", ErrRestletError, resp.Header.Status)
	}

	return fmt.Errorf("%w: [%s] %s", ErrRestletError, errBody.ErrorCode, errBody.ErrorMessage)
}

// interpretError handles non-200 HTTP responses (network errors, auth failures, etc.).
// The RESTlet itself always returns 200, but the NS platform may return 401/403/etc.
func interpretError(resp []byte, statusCode int) error {
	switch {
	case statusCode == 401 || statusCode == 403:
		return fmt.Errorf("%w: HTTP %d", common.ErrAccessToken, statusCode)
	case statusCode >= 500:
		return fmt.Errorf("%w: HTTP %d", common.ErrServer, statusCode)
	case statusCode >= 400:
		return fmt.Errorf("%w: HTTP %d: %s", common.ErrCaller, statusCode, string(resp))
	default:
		return nil
	}
}
