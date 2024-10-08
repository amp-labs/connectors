package dynamicscrm

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// Write data will be used to Create or Update entity.
// Return: common.WriteResult, where only the Success flag will be set.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var resource string

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means
		// that we are extending 'List' resource and creating a new record
		write = c.JSON.Post
		resource = config.ObjectName
	} else {
		// only patch is supported for updating 'Single' resource
		write = c.JSON.Patch
		// resource id is passed via brackets in OData spec
		resource = fmt.Sprintf("%s(%s)", config.ObjectName, config.RecordId)
	}

	url, err := c.getURL(resource)
	if err != nil {
		return nil, err
	}

	// Neither Post nor Patch return any response data on successful completion
	// Both complete with 204 NoContent
	_, err = write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
	}, nil
}
