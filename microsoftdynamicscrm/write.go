package microsoftdynamicscrm

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

type writeMethod func(context.Context, string, any) (*common.JSONHTTPResponse, error)

// Write data will be used to Create or Update entity.
// Return: common.WriteResult, where only the Success flag will be set.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	url := c.getURL(config.ObjectName)

	var write writeMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means
		// that we are extending 'List' resource and creating a new record
		write = c.post
	} else {
		// only patch is supported for updating 'Single' resource
		write = c.patch
		// resource id is passed via brackets in OData spec
		url = fmt.Sprintf("%s(%s)", url, config.RecordId)
	}

	// Neither Post nor Patch return any response data on successful completion
	// Both complete with 204 NoContent
	_, err := write(ctx, url, config.RecordData)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
	}, nil
}
