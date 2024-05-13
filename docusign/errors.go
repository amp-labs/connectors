package docusign

import (
	"errors"
)

var ErrMissingClient = errors.New("JSON http client not set")

func (c *Connector) HandleError(err error) error {
	return err
}
