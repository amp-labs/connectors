package intercom

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

type writeMethod func(context.Context, string, any) (*common.JSONHTTPResponse, error)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write writeMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means
		// that we are extending 'List' resource and creating a new record
		write = c.post
	} else {
		// only put is supported for updating 'Single' resource
		write = c.put

		url.AddPath(config.RecordId)
	}

	res, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	if res == nil || res.Body == nil {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return constructWriteResult(res.Body)
}

func constructWriteResult(body *ajson.Node) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(body).StrWithDefault("id", "")
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
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
