package instantly

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Delete removes object. As of now only removal of Tags are allowed because
// deletion of other object types require a request payload to be added
// c.Client.Delete does not yet support this.
func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if len(config.RecordId) == 0 {
		return nil, common.ErrMissingRecordID
	}

	if !supportedObjectsByDelete.Has(config.ObjectName) {
		// Removing tags is the only to be supported at this time.
		// https://developer.instantly.ai/tags/delete-a-tag
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getURL("custom-tag", config.RecordId)
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
