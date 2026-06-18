package mail

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (a *Adapter) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	obj, err := lookupWriteObject(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := a.objectURL(obj.path, obj.accountScoped)
	if err != nil {
		return nil, err
	}

	write := a.Client.Post
	if config.IsUpdate() {
		url.AddPath(config.RecordId)
		write = a.Client.Put
	}

	resp, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return parseWriteResult(resp, obj)
}

func parseWriteResult(resp *common.JSONHTTPResponse, obj writeDescriptor) (*common.WriteResult, error) {
	node, ok := resp.Body()
	if !ok {
		// Some endpoints return an empty body on success.
		return &common.WriteResult{Success: true}, nil
	}

	data, err := jsonquery.New(node).ObjectOptional("data")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return &common.WriteResult{Success: true}, nil
	}

	// recordIdKey may be numeric (e.g. folderId), so read it as text.
	recordID, err := jsonquery.New(data).TextWithDefault(obj.recordIdKey, "")
	if err != nil {
		return nil, err
	}

	recordData, err := jsonquery.Convertor.ObjectToMap(data)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     recordData,
	}, nil
}
