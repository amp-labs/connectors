// nolint
package attio

import "errors"

var ErrEmptyResultResponse = errors.New("writing reponded with an empty result")

type writeResponse struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
}
