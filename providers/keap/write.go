package keap

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
		// Update is done either by PUT or PATCH. There is no object present in both sets.
		if supportedObjectsByUpdatePUT[c.Module.ID].Has(config.ObjectName) {
			write = c.Client.Put

			url.AddPath(config.RecordId)
		}

		if supportedObjectsByUpdatePATCH[c.Module.ID].Has(config.ObjectName) {
			write = c.Client.Patch

			url.AddPath(config.RecordId)
		}
	}

	if write == nil {
		// No supported REST operation was found for current object.
		return nil, common.ErrOperationNotSupportedForObject
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
	return constructWriteResult(config, c.Module.ID, body)
}

func constructWriteResult(
	config common.WriteParams, moduleID common.ModuleID, body *ajson.Node,
) (*common.WriteResult, error) {
	identifierHolder := body
	// Object "files" is the only exception where the identifier is located, it is nested.
	if config.ObjectName == objectNameFiles {
		var err error
		// Identifier is nested under "file_descriptor" object.
		identifierHolder, err = jsonquery.New(body).Object("file_descriptor", false)
		if err != nil {
			return nil, err
		}
	}

	writeIdentifier := objectNameToWriteResponseIdentifier[moduleID].Get(config.ObjectName)

	recordID, err := jsonquery.New(identifierHolder).TextWithDefault(writeIdentifier, "")
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
