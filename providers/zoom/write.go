package zoom

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Each object has different fields that represent the record id.
// This map is used to get the record id field for each object.
var recordIdPaths = map[string]string{ //nolint:gochecknoglobals
	ObjectNameUser:          "id",
	ObjectNameContactGroup:  "group_id",
	ObjectNameGroup:         "id",
	objectNameTrackingField: "id",
}

func (c *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) { //nolint:funlen
	err := config.ValidateParams()
	if err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	var write common.WriteMethod

	url, err := c.getURL(config.ObjectName)
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

	recordIdPath := recordIdPaths[config.ObjectName]

	// write response with payload
	return constructWriteResult(body, recordIdPath)
}

func constructWriteResult(body *ajson.Node, recordIdLocation string) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(body).Str(recordIdLocation, false)
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
