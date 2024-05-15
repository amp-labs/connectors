package interpreter

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
)

func DefaultStatusCodeMappingToErr(res *http.Response, body []byte) error { // nolint:cyclop
	switch res.StatusCode {
	case http.StatusBadRequest:
		return common.ErrBadRequest
	case http.StatusUnauthorized:
		return common.ErrAccessToken
	case http.StatusForbidden:
		return common.ErrForbidden
	case http.StatusNotFound:
		return common.ErrBadRequest // TODO more specific error
	case http.StatusMethodNotAllowed:
		return common.ErrBadRequest // TODO more specific error
	case http.StatusPreconditionFailed:
		return common.ErrBadRequest // TODO more specific error
	case http.StatusRequestEntityTooLarge:
		return common.ErrBadRequest // TODO more specific error
	case http.StatusTooManyRequests:
		return common.ErrLimitExceeded
	case http.StatusNotImplemented:
		return common.ErrNotImplemented
	case http.StatusServiceUnavailable:
		return common.ErrServer
	default:
		return common.InterpretError(res, body)
	}
}
