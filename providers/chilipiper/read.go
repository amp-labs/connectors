package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (conn *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := conn.buildURL(config.ObjectName, readpageSize)
	if err != nil {
		return nil, err
	}

	// Check if we're reading Next Page of Records.
	if len(config.NextPage) > 0 {
		url = config.NextPage.String()
	}

	resp, err := conn.Client.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.GetRecordsUnderJSONPath("results"),
		nextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}
