// nolint
package attio

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Write creates/updates records in marketo. Write currently supports operations to the leads API only.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means
		// that we are extending 'List' resource and creating a new record
		write = c.Client.Post
	} else {
		// only put is supported for updating 'Single' resource
		write = c.Client.Patch

		url.AddPath(config.RecordId)
	}

	res, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	resp, err := common.UnmarshalJSON[writeResponse](res)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, ErrEmptyResultResponse
	}
	if res.Code == 200 {
		resp.Success = true
	}

	return &common.WriteResult{
		Success: resp.Success,
		Data:    resp.Data,
	}, nil
}
