package freshdesk

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
)

func (conn *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	ctx = logging.With(ctx, "connector", "freshdesk")
	write := conn.JSONHTTPClient().Post

	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !writeSupportedObjects.Has(config.ObjectName) {
		return nil, common.ErrObjectNotSupported
	}

	url, err := conn.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		url.AddPath(config.RecordId)

		write = conn.JSONHTTPClient().Put
	}

	resp, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return constructWriteResponse(ctx, config.ObjectName, resp), nil
}

func constructWriteResponse(ctx context.Context, objectName string, resp *common.JSONHTTPResponse) *common.WriteResult {
	res, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		logging.Logger(ctx).Warn("failed to unmarshal response: ", "error: ", err.Error(), "objectName: ", objectName)

		return &common.WriteResult{
			Success: true,
		}
	}

	data := *res

	recordId, ok := data["id"].(float64)
	if !ok {
		logging.Logger(ctx).Warn("failed to cast 'id' field to float64", "objectName: ", objectName)

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
