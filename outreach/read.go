package outreach

import (
	"context"
	"net/url"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		res *common.JSONHTTPResponse
		err error
	)

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// The NextPage URL has all the necessary parameters.
		res, err = c.get(ctx, config.NextPage)
		if err != nil {
			return nil, err
		}

	} else {
		fullURL, err := url.JoinPath(c.BaseURL, config.ObjectName)
		if err != nil {
			return nil, err
		}

		// Testing pagination
		// fullURL = fullURL + "?page[limit]=1"

		res, err = c.get(ctx, fullURL)
		if err != nil {
			return nil, err
		}
	}

	return common.ParseResult(res, getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshaledData,
		config.Fields)
}
