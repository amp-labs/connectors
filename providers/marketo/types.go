package marketo

import (
	"errors"
	"fmt"
)

var ErrEmptyResultResponse = errors.New("writing reponded with an empty result")

type writeResponse struct {
	Result  []map[string]any `json:"result"`
	Success bool             `json:"success"`
	Errors  []map[string]any `json:"errors"`
}

// IdResponses represents a list of objects that uses `id` as a unique field in the response.
var IdResponses = []string{"leads", "companies", "salespersons"} //nolint:gochecknoglobals

// marketoGUIDResponses represents a list of objects that uses `marketoGUID` as the unique field in the response.
var marketoGUIDResponses = []string{"namedAccountLists", "namedaccounts", "opportunities"} //nolint:gochecknoglobals

func constructErrMessage(a any) (string, error) {
	s, ok := a.([]map[string]any)
	if !ok {
		return "", errors.New("failed to convert the response message to an error type") // nolint:goerr113
	}

	return fmt.Sprint(s[0]["reasons"]), nil
}
