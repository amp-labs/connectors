package restlet

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrRestletError              = errors.New("restlet error")
	ErrUnsupportedFilterOperator = errors.New("unsupported filter operator")
)

func parseRestletError(resp *restletResponse) error {
	var errBody restletErrorBody
	if err := json.Unmarshal(resp.Body, &errBody); err != nil {
		return fmt.Errorf("%w: status=%s", ErrRestletError, resp.Header.Status)
	}

	return fmt.Errorf("%w: [%s] %s", ErrRestletError, errBody.ErrorCode, errBody.ErrorMessage)
}
