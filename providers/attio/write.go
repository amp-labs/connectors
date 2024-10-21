// nolint
package attio

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
)

var ErrEmptyResultResponse = errors.New("writing reponded with an empty result")

type writeResponse struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data"`
}

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

	resp, err := common.UnmarshalJSON[writeResponse](res)
	if err != nil {
		return nil, err
	}

	objectIdData, ok := resp.Data["id"]
	if !ok {
		return nil, jsonquery.ErrKeyNotFound
	}

	recordID, err := GetRecordID(config.ObjectName, objectIdData.(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     resp.Data,
		Errors:   nil,
	}, nil
}

func GetRecordID(objName string, data map[string]interface{}) (string, error) {
	obj := naming.NewSingularString(objName)

	if value, ok := data[obj.String()+"_id"]; ok {
		return value.(string), nil
	}

	return "", jsonquery.ErrKeyNotFound
}
