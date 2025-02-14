package front

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/spyzhov/ajson"
)

// Write only supports creating Calls.
func (conn *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	write := conn.Client.Post

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := conn.getBaseAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		if !supportsPatching(config.ObjectName) {
			return nil, common.ErrObjectNotSupported
		}

		url.AddPath(config.RecordId)
		write = conn.Client.Patch
	}

	if !supportsCreation(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
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

	return constructWriteResult(body)
}

func constructWriteResult(body *ajson.Node) (*common.WriteResult, error) {
	recordID, err := jsonquery.New(body).Str("id", false)
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
