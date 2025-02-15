package asana

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	err := config.ValidateParams()
	if err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	var write common.WriteMethod

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) == 0 {
		// writing to the entity without id means creating a new record.
		write = c.Client.Post
	} else {
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

	// write response with payload
	return constructWriteResult(body)
}

func constructWriteResult(body *ajson.Node) (*common.WriteResult, error) {
	nested, err := jsonquery.New(body).Object("data", false)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(nested).StrWithDefault("gid", "")
	if err != nil {
		return nil, err
	}

	response, err := jsonquery.Convertor.ObjectToMap(nested)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     response,
	}, nil
}
