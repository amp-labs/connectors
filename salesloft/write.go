package salesloft

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

type writeMethod func(context.Context, string, any) (*common.JSONHTTPResponse, error)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	url := c.getURL(config.ObjectName)

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

	// TODO investigate that all endpoints "wrap" payload under `data` key
	nested, err := common.JSONManager.GetNestedNode(res.Body, []string{"data"})
	if err != nil {
		return nil, err
	}

	rawID, err := common.JSONManager.GetInteger(nested, "id", true)
	if err != nil {
		return nil, err
	}

	recordID := ""
	if rawID != nil {
		// optional
		recordID = strconv.FormatInt(*rawID, 10)
	}

	data, err := common.JSONManager.ObjToMap(nested)
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
