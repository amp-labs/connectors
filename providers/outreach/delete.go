package outreach

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Delete removes a record from Outreach.
// The Outreach API returns HTTP 204 No Content on successful deletion.
func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(config.RecordId)

	// 204 NoContent is expected from Outreach API
	_, err = c.Client.Delete(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
