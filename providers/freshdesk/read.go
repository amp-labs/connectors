package freshdesk

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := conn.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	resp, err := conn.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		extractRecords,
		nextRecordsURL(resp),
		common.GetMarshaledData,
		config.Fields,
	)
}
