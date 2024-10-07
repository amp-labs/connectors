// nolint
package attio

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
)

var ErrEmptyResultResponse = errors.New("writing reponded with an empty result")

type writeResponse struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
}

// Write creates/updates records in attio.
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
		// writing to the entity without id means creating a new record
		write = c.Client.Post
	} else {
		// updating resource by patch method
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

	if res.Code == 200 {
		resp.Success = true
	}

	return &common.WriteResult{
		Success: resp.Success,
		Data:    resp.Data,
	}, nil
}
