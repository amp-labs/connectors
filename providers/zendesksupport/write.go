package zendesksupport

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
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
		write = c.Client.Post
	} else {
		// only put is supported for updating 'Single' resource
		write = c.Client.Put

		url.AddPath(config.RecordId)
	}

	res, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return constructWriteResult(config, body)
}

func constructWriteResult(config common.WriteParams, body *ajson.Node) (*common.WriteResult, error) {
	nested, err := jsonquery.New(body).Object(config.ObjectName, true)
	if err != nil {
		return nil, err
	}

	if nested == nil {
		// Field should be in singular form. Either one will be matched.
		// This one is NOT optional.
		nested, err = jsonquery.New(body).Object(
			naming.NewSingularString(config.ObjectName).String(),
			false,
		)
		if err != nil {
			return nil, err
		}
	}
	// nested node now must be not null, carry on

	rawID, err := jsonquery.New(nested).Integer("id", true)
	if err != nil {
		return nil, err
	}

	recordID := ""
	if rawID != nil {
		// optional
		recordID = strconv.FormatInt(*rawID, 10)
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
