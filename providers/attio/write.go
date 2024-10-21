// nolint
package attio

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/spyzhov/ajson"
)

var ErrEmptyResultResponse = errors.New("writing reponded with an empty result")

// Write creates/updates records in attio.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means creating a new record.
		write = c.Client.Post
	} else {
		// updating resource by patch method.
		write = c.Client.Patch

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

	// Write response has a reference to the resource but no payload data.
	return constructWriteResult(config.ObjectName, body)
}

func constructWriteResult(objName string, body *ajson.Node) (*common.WriteResult, error) {
	obj := naming.NewSingularString(objName)

	objectResponse, err := jsonquery.New(body).Object("data", false)
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(objectResponse, "id").Str(obj.String()+"_id", false)
	if err != nil {
		return nil, err
	}

	response, err := jsonquery.Convertor.ObjectToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *recordID,
		Errors:   nil,
		Data:     response,
	}, nil
}
