package gong

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

	if config.Fields != nil {
		fields = config.Fields

	} else {
		fields = []string{"*"}
	}

	fullURL, err := url.JoinPath(c.BaseURL, config.ObjectName)
	if err != nil {
		return nil, err
	}

	res, err = c.get(ctx, fullURL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res, getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshaledData,
		fields,
	)
}
