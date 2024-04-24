package gong

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) get(ctx context.Context, url string) (*common.JSONHTTPResponse, error) {
	res, err := c.Client.Get(ctx, url)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return res, nil
}
