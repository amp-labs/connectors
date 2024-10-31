package closecrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Write creates/updates records in closecrm.
//
// doc: https://developer.close.com/resources/leads/#create-a-new-lead
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	var write common.WriteMethod

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}
	// sets post as default
	write = c.Client.Post

	// prepares the updating data request.
	if len(config.RecordId) > 0 {
		url = url.AddPath(config.RecordId)

		write = c.Client.Put
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

	return constructWriteResult(body)
}

func constructWriteResult(node *ajson.Node) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(node).Str("id", false)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(node)
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
