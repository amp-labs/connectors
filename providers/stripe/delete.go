package stripe

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(config.RecordId)

	// 200 OK is expected
	_, err = c.Client.Delete(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
