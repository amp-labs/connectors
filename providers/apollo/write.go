package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/spyzhov/ajson"
)

// Write creates/updates records in apolllo.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	var write common.WriteMethod

	url, err := c.getAPIURL(config.ObjectName, writeOp)
	if err != nil {
		return nil, err
	}
	// sets post as default
	write = c.Client.Post

	// prepares the updating data request.
	if len(config.RecordId) > 0 {
		url = url.AddPath(config.RecordId)

		write = c.Client.Patch
	}

	json, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	body, ok := json.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	return constructWriteResult(body, config.ObjectName)
}

func constructWriteResult(body *ajson.Node, objName string) (*common.WriteResult, error) {
	// API Response contains a json object having a singular objectName key with the
	// created/updated details in it.
	obj := naming.NewSingularString(objName)

	respObject, err := jsonquery.New(body).Object(obj.String(), false)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(respObject).StrWithDefault("id", "")
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
