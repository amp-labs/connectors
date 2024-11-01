package constantcontact

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Write only supports creating Calls.
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
		if !supportedObjectsByCreate.Has(config.ObjectName) {
			return nil, common.ErrOperationNotSupportedForObject
		}

		write = c.Client.Post
	} else {
		if !supportedObjectsByUpdate.Has(config.ObjectName) {
			return nil, common.ErrOperationNotSupportedForObject
		}

		write = c.Client.Put
		if config.ObjectName == objectNameEmailCampaigns {
			write = c.Client.Patch
		}

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
	idFieldName := objectNameToWriteResponseIdentifier.Get(config.ObjectName)

	recordID, err := jsonquery.New(body).Str(idFieldName, false)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
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
