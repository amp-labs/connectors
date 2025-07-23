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

	// we're using the link response header to construct the next page url.
	linkHeader := resp.Headers.Get("link")

	return common.ParseResult(
		resp,
		common.ExtractRecordsFromPath(""), // we're reading the current node.
		nextRecordsURL(linkHeader),
		common.GetMarshaledData,
		config.Fields,
	)
}
