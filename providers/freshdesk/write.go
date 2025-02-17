package freshdesk

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	write := conn.Client.Post

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !objectWriteSupported(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
	}

	url, err := conn.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		url.AddPath(config.RecordId)

		write = conn.Client.Put
	}

	resp, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return constructWriteResponse(resp), nil
}

func constructWriteResponse(resp *common.JSONHTTPResponse) *common.WriteResult {
	res, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return &common.WriteResult{
			Success: true,
		}
	}

	data := *res

	recordId, ok := data["id"].(float64)
	if !ok {
		return &common.WriteResult{
			Success: true,
		}
	}

	return &common.WriteResult{
		RecordId: strconv.Itoa(int(recordId)),
		Success:  true,
		Data:     data,
	}
}
