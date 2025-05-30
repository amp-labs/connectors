package klaviyo

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByDelete[common.ModuleRoot].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getDeleteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.AddPath(config.RecordId)

	// 200 OK is expected
	_, err = c.Client.Delete(ctx, url.String(), c.revisionHeader())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
