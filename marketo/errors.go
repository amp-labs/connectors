package marketo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Respose struct {
	Errors    []Error          `json:"errors"`
	Result    []map[string]any `json:"result"`
	RequestID string           `json:"requestId"`
	Success   bool             `json:"success"`
}

// checkResponseLeverErr reports wheather the response level error is available or not.
// If available, returns the error code as well.
func checkResponseLeverErr(body []byte) (bool, int, error) {
	var resp Respose
	if err := json.Unmarshal(body, &resp); err != nil {
		return false, 0, err
	}

	code, err := strconv.Atoi(resp.Errors[0].Code)
	if err != nil {
		return false, 0, err
	}

	return len(resp.Errors) > 0, code, nil
}

// InterpretError interprets the given HTTP response (in a fairly straightforward
// way) and returns an error that can be handled by the caller.
func interpretError(res *http.Response, body []byte) error { //nolint:cyclop
	// A must check.
	if res.StatusCode < 200 || res.StatusCode > 299 {
		erroneous, code, err := checkResponseLeverErr(body)
		if err != nil {
			return err
		}
		fmt.Println("Erroneous: ", erroneous)

		// If response is 200 OK, but erroneous, we update the status code & continue with the switch cases.
		if erroneous {
			statusCode := statusCodeMap(code)
			res.StatusCode = statusCode
		} else {
			return nil
		}
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		// Access token invalid, refresh token and retry
		return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrAccessToken, string(body)))
	case http.StatusForbidden:
		// Forbidden, not retryable
		return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrForbidden, string(body)))
	case http.StatusNotFound:
		// Semantics are debatable (temporarily missing vs. permanently gone), but for now treat this as a retryable error
		return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrRetryable, string(body)))
	case http.StatusTooManyRequests:
		// Too many requests, retryable
		return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrRetryable, string(body)))
	}

	if res.StatusCode >= 400 && res.StatusCode < 500 {
		return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrCaller, string(body)))
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrServer, string(body)))
	}

	return common.NewHTTPStatusError(res.StatusCode, fmt.Errorf("%w: %s", common.ErrUnknown, string(body)))
}

// statusCodeMap maps the erroneous response from marketo, with a valid http status code.
// The response body can be sent as is.
// https://experienceleague.adobe.com/en/docs/marketo-developer/marketo/rest/error-codes
func statusCodeMap(code int) int { //nolint:funlen,cyclop
	switch code {
	case 502: //nolint:gomnd
		return http.StatusBadGateway
	case 601: //nolint:gomnd
		return http.StatusUnauthorized
	case 602: //nolint:gomnd
		return http.StatusUnauthorized // token Expired
	case 603: //nolint:gomnd
		return http.StatusForbidden
	case 604: //nolint:gomnd
		return http.StatusRequestTimeout
	case 605: //nolint:gomnd
		return http.StatusMethodNotAllowed
	case 606: //nolint:gomnd
		return http.StatusTooManyRequests
	case 607: //nolint:gomnd
		return http.StatusTooManyRequests
	case 608: //nolint:gomnd
		return http.StatusServiceUnavailable
	case 609: //nolint:gomnd
		return http.StatusBadRequest
	case 610: //nolint:gomnd
		return http.StatusNotFound
	case 611: //nolint:gomnd
		return http.StatusInternalServerError
	case 612: //nolint:gomnd
		return http.StatusUnsupportedMediaType
	case 613: //nolint:gomnd
		return http.StatusBadRequest
	case 614: //nolint:gomnd
		return http.StatusNotFound
	case 615: //nolint:gomnd
		return http.StatusTooManyRequests
	case 616: //nolint:gomnd
		return http.StatusForbidden
	case 701: //nolint:gomnd
		return http.StatusBadRequest
	case 702: //nolint:gomnd
		return http.StatusNotFound
	case 703: //nolint:gomnd
		return http.StatusForbidden
	case 704: //nolint:gomnd
		return http.StatusBadRequest
	case 709: //nolint:gomnd
		return http.StatusConflict
	case 710: //nolint:gomnd
		return http.StatusNotFound
	case 711: //nolint:gomnd
		return http.StatusBadRequest
	case 712: //nolint:gomnd
		return http.StatusBadRequest
	case 713: //nolint:gomnd
		return http.StatusServiceUnavailable
	case 714: //nolint:gomnd
		return http.StatusNotFound
	case 718: //nolint:gomnd
		return http.StatusNotFound
	case 719: //nolint:gomnd
		return http.StatusRequestTimeout
	default:
		return code
	}
}
