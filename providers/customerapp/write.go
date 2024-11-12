package customerapp

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

	path := ObjectNameToWritePath.Get(config.ObjectName)

	url, err := c.getURL(path)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod

	if len(config.RecordId) == 0 {
		if !supportedObjectsByCreate.Has(config.ObjectName) {
			return nil, common.ErrOperationNotSupportedForObject
		}

		write = c.Client.Post
		if config.ObjectName == objectNameSnippets {
			// https://docs.customer.io/api/app/#operation/listSnippets
			// Snippets are create and updated via PUT.
			write = c.Client.Put
		}
	} else {
		if !supportedObjectsByUpdate.Has(config.ObjectName) {
			return nil, common.ErrOperationNotSupportedForObject
		}

		write = c.Client.Put

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
	return constructWriteResult(config, body)
}

func constructWriteResult(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
	fieldName := ObjectNameToWriteResponseField.Get(config.ObjectName)

	nested, err := jsonquery.New(body).Object(fieldName, false)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(nested).TextWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(nested)
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
