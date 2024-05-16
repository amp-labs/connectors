package microsoftdynamicscrm

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if len(config.RecordId) == 0 {
		return nil, common.ErrMissingRecordID
	}

	// resource id must be enclosed in brackets
	url := c.getURL(fmt.Sprintf("%v(%v)", config.ObjectName, config.RecordId))

	// 204 NoContent is expected
	_, err := c.delete(ctx, url)
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
