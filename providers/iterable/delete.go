package iterable

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByDelete.Has(config.ObjectName) {
		// Removing tags is the only to be supported at this time.
		// https://developer.instantly.ai/tags/delete-a-tag
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getDeleteURL(config.ObjectName, config.RecordId)
	if err != nil {
		return nil, err
	}

	// 200 OK is expected
	_, err = c.Client.Delete(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
