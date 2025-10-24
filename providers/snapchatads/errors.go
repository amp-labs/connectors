package snapchatads

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var errorFormats = interpreter.NewFormatSwitch( // nolint:gochecknoglobals
	[]interpreter.FormatTemplate{
		{
			MustKeys: nil,
			Template: func() interpreter.ErrorDescriptor { return &ResponseError{} },
		},
	}...,
)

var errObjNotFound = errors.New("object not found")

type ResponseError struct {
	RequestStatus  string `json:"request_status"`  //nolint:tagliatelle
	RequestId      string `json:"request_id"`      //nolint:tagliatelle
	DebugMessage   string `json:"debug_message"`   //nolint:tagliatelle
	DisplayMessage string `json:"display_message"` //nolint:tagliatelle
	ErrorCode      string `json:"error_code"`      //nolint:tagliatelle
}

func (r ResponseError) CombineErr(base error) error {
	return fmt.Errorf("%w: [%v] %v", base, r.ErrorCode, r.DisplayMessage)
}

func checkErrorInResponse(resp *common.JSONHTTPResponse, objectName string) (bool, error) {
	body, ok := resp.Body() // nolint:varnamelen
	if !ok {
		return false, common.ErrEmptyJSONHTTPResponse
	}

	objectResponse, err := jsonquery.New(body).ArrayRequired(objectName)
	if err != nil {
		return false, err
	}

	res, err := jsonquery.Convertor.ArrayToMap(objectResponse)
	if err != nil {
		return false, err
	}

	if len(res) != 0 {
		return true, nil
	}

	return false, nil
}

func responseHandler(resp *common.JSONHTTPResponse, objName string) (*common.JSONHTTPResponse, error) { //nolint:cyclop
	if resp.Code >= 200 && resp.Code <= 299 {
		erroneous, err := checkErrorInResponse(resp, objName)
		if err != nil {
			return nil, err
		}

		// If response is 200 OK with error body, convert into 400 statuscode.
		if erroneous {
			statusCode := http.StatusBadRequest
			resp.Code = statusCode
		}
	}

	return resp, nil
}
