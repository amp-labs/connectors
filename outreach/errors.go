package outreach

import (
	"errors"

	"github.com/amp-labs/connectors/common"
)

var (
	ErrMissingClient = errors.New("JSON http client not set")
	ErrNotArray      = errors.New("results data is not an array")
	ErrNotObject     = errors.New("record is not an object")
	ErrNotString     = errors.New("next is not a string")
)

func (c *Connector) HandleError(err error) error {
	switch {
	case errors.Is(err, common.ErrAccessToken):
		// Retryable, so just log and retry
		// TODO: Retry
		return err
	case errors.Is(err, common.ErrRetryable):
		// TODO: Retry
		return err
	case errors.Is(err, common.ErrApiDisabled):
		fallthrough
	case errors.Is(err, common.ErrForbidden):
		fallthrough
	default:
		// Anything else is a permanent error
		return err
	}
}
