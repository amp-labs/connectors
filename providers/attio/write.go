package attio

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var ErrEmptyResultResponse = errors.New("writing responded with an empty result")

// Write creates/updates records in attio.
func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.buildWriteURL(config)
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		// writing to the entity without id means creating a new record.
		write = c.Client.Post
	} else {
		// updating resource by patch method.
		write = c.Client.Put

		if supportWriteObjects.Has(config.ObjectName) {
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

	// Write response has a reference to the resource but no payload data.
	return constructWriteResult(config.ObjectName, body)
}

func (c *Connector) buildWriteURL(config common.WriteParams) (*urlbuilder.URL, error) {
	// supportWriteObjects represents non-standard/custom objects.
	if supportWriteObjects.Has(config.ObjectName) {
		return c.getApiURL(config.ObjectName)
	}

	url, err := c.getObjectWriteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	return url, nil
}

func constructWriteResult(objName string, body *ajson.Node) (*common.WriteResult, error) {
	var obj naming.SingularString

	if supportWriteObjects.Has(objName) {
		obj = naming.NewSingularString(objName)
	} else {
		obj = naming.NewSingularString("record")
	}

	objectResponse, err := jsonquery.New(body).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	recordID, err := jsonquery.New(objectResponse, "id").StringRequired(obj.String() + "_id")
	if err != nil {
		return nil, err
	}

	response, err := jsonquery.Convertor.ObjectToMap(objectResponse)
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
