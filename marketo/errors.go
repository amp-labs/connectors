package marketo

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// retrieveInternalCode returns the first error code in marketo http response.
func retrieveInternalCode(root *ajson.Node) (int, error) {
	var errD Error
	Errors, err := jsonquery.New(root).Array("errors", true)
	if err != nil {
		return 0, err
	}

	if err := json.Unmarshal(Errors[0].Source(), &errD); err != nil {
		return 0, err
	}

	code, err := strconv.Atoi(errD.Code)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// checkResponseLeverErr reports wheather the response level error is available.
func checkResponseLeverErr(root *ajson.Node) (bool, error) {
	size, err := jsonquery.New(root).ArraySize("errors")
	if err != nil {
		return false, err
	}

	return size > 0, nil
}

func ErrorHandler(resp *http.Response)

// statusCodeMap maps the erroneous response from marketo, with a valid http status code.
// The response body can be sent as is.
// https://experienceleague.adobe.com/en/docs/marketo-developer/marketo/rest/error-codes
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
