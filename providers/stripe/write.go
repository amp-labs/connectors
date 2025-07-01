package stripe

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod

	if len(config.RecordId) == 0 {
		write = c.Client.Post
	} else {
		write = c.Client.Post

		url.AddPath(config.RecordId)
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
	recordID, err := jsonquery.New(node).StringRequired("id")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}
