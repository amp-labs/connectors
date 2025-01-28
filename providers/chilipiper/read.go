package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	readURL, err := conn.buildReadURL(config)
	if err != nil {
		return nil, err
	}

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
