package servicenow

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	write := c.Client.Post

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		url.AddPath(config.RecordId)

		write = c.Client.Patch
	}

	resp, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	return constructWriteResult(body)
}

func constructWriteResult(body *ajson.Node) (*common.WriteResult, error) {
	result, err := jsonquery.New(body).Object("result", false)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(result)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success: true,
		Errors:  nil,
		Data:    data,
	}, nil
}
