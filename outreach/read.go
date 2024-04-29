package outreach

import (
	"context"
	"net/url"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		res    *common.JSONHTTPResponse
		err    error
		fields []string
	)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// The NextPage URL has all the necessary parameters.
		res, err = c.get(ctx, config.NextPage.String())
		if err != nil {
			return nil, err
		}
	} else {
		fullURL, err := url.JoinPath(c.BaseURL, config.ObjectName)
		if err != nil {
			return nil, err
		}

		res, err = c.get(ctx, fullURL)
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(res, getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshaledData,
		fields,
	)
}
