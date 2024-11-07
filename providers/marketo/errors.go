package marketo

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/httputils"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Errors    []Error          `json:"errors"`
	Result    []map[string]any `json:"result"`
	RequestID string           `json:"requestId"`
	Success   bool             `json:"success"`
}

// checkResponseLeverErr reports wheather the response level error is available or not.
// If available, returns the error code as well.
func checkResponseLeverErr(body []byte) (bool, int, error) {
	var resp Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return false, 0, err
	}

	if len(resp.Errors) == 0 {
		return false, 0, nil
	}

	code, err := strconv.Atoi(resp.Errors[0].Code)
	if err != nil {
		return false, 0, err
	}

	return len(resp.Errors) > 0, code, nil
}

func isSuccessfulResponse(resp *http.Response, body []byte) (bool, error) {
	if success := httputils.IsStatus2XX(resp); success {
		return true, nil
	}

	erroneous, _, err := checkResponseLeverErr(body)
	if err != nil {
		return false, err
	}

	return !erroneous, nil
}

func interpretError(res *http.Response, body []byte) error {
	erroneous, code, err := checkResponseLeverErr(body)
	if err != nil {
		return err
	}

	// If response is 200 OK, but erroneous, we update the status code
	if erroneous {
		statusCode := statusCodeMap(code)
		res.StatusCode = statusCode
	}

	return common.InterpretError(res, body)
}

// statusCodeMap maps the erroneous response from marketo, with a valid http status code.
// The response body can be sent as is.
// https://experienceleague.adobe.com/en/docs/marketo-developer/marketo/rest/error-codes
//
// nolint
func statusCodeMap(code int) int {
	switch code {
	case 502:
		return http.StatusBadGateway
	case 601:
		return http.StatusUnauthorized
	case 602:
		return http.StatusUnauthorized // token Expired
	case 603:
		return http.StatusForbidden
	case 604:
		return http.StatusRequestTimeout
	case 605:
		return http.StatusMethodNotAllowed
	case 606:
		return http.StatusTooManyRequests
	case 607:
		return http.StatusTooManyRequests
	case 608:
		return http.StatusServiceUnavailable
	case 609:
		return http.StatusBadRequest
	case 610:
		return http.StatusNotFound
	case 611:
		return http.StatusInternalServerError
	case 612:
		return http.StatusUnsupportedMediaType
	case 613:
		return http.StatusBadRequest
	case 614:
		return http.StatusNotFound
	case 615:
		return http.StatusTooManyRequests
	case 616:
		return http.StatusForbidden
	case 701:
		return http.StatusBadRequest
	case 702:
		return http.StatusNotFound
	case 703:
		return http.StatusForbidden
	case 704:
		return http.StatusBadRequest
	case 709:
		return http.StatusConflict
	case 710:
		return http.StatusNotFound
	case 711:
		return http.StatusBadRequest
	case 712:
		return http.StatusBadRequest
	case 713:
		return http.StatusServiceUnavailable
	case 714:
		return http.StatusNotFound
	case 718:
		return http.StatusNotFound
	case 719:
		return http.StatusRequestTimeout
	default:
		return code
	}
}
