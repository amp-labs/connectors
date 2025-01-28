package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	path, err := supportsRead(config.ObjectName)
	if err != nil {
		return nil, err
	}

	readURL, err := conn.buildReadURL(path)
	if err != nil {
		return nil, err
	}

	readURL.WithQueryParam(pageSizeKey, pageSize)

	resp, err := conn.Client.Get(ctx, readURL.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.GetRecordsUnderJSONPath("results"),
		nextRecordsURL(readURL),
		common.GetMarshaledData,
		config.Fields,
	)

}
