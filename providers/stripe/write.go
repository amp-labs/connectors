package stripe

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getWriteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod

	if len(config.RecordId) == 0 {
		if supportedObjectsByCreate[c.Module.ID].Has(config.ObjectName) {
			write = c.Client.Post
		}
	} else {
		if supportedObjectsByUpdate[c.Module.ID].Has(config.ObjectName) {
			write = c.Client.Post

			url.AddPath(config.RecordId)
		}
	}

	if write == nil {
		// No supported REST operation was found for current object.
		return nil, common.ErrOperationNotSupportedForObject
	}

	res, err := write(ctx, url.String(), config.RecordData, common.HeaderFormURLEncoded)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return constructWriteResult(body)
}

func constructWriteResult(node *ajson.Node) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(node).Str("id", false)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}
