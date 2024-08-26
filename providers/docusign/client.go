package docusign

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// JSONHTTPClient returns the underlying JSON HTTP client.
func (c *Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return c.Client
}

func (c *Connector) HTTPClient() *common.HTTPClient {
	return c.Client.HTTPClient
}

func (c *Connector) Close() error {
	return nil
}

func (c *Connector) String() string {
	return c.Provider() + ".Connector"
}

func (c *Connector) get(ctx context.Context, url string) (*common.JSONHTTPResponse, error) {
	res, err := c.Client.Get(ctx, url)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return res, nil
}

func (c *Connector) HandleError(err error) error {
	return err
}
