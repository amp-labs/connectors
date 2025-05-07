package zoom

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) { //nolint:funlen
	err := config.ValidateParams()
	if err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite[c.moduleID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	var write common.WriteMethod

	url, err := c.getWriteURL(config.ObjectName)
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

	recordIdPath := objectNameToWriteResponseIdentifier[c.moduleID].Get(config.ObjectName)

	// write response with payload
	return constructWriteResult(body, recordIdPath)
}

func constructWriteResult(body *ajson.Node, recordIdLocation string) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(body).StringRequired(recordIdLocation)
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
