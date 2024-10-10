package pipeliner

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

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means
		// that we are extending 'List' resource and creating a new record
		write = c.Clients.JSON.Post
	} else {
		// only put is supported for updating 'Single' resource
		write = c.Clients.JSON.Patch

		url.AddPath(config.RecordId)
	}

	res, err := write(ctx, url.String(), config.RecordData)
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

func constructWriteResult(body *ajson.Node) (*common.WriteResult, error) {
	success, err := jsonquery.New(body).Bool("success", false)
	if err != nil {
		return nil, err
	}

	nested, err := jsonquery.New(body).Object("data", false)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(nested).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(nested)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  *success,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}
