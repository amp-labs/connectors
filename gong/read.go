package gong

import (
	"context"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		res *common.JSONHTTPResponse

		fields []string
	)

	fullURL, err := url.JoinPath(c.BaseURL, c.APIModule.Version, config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.NextPage) != 0 { // not the first page, add a cursor

		fullURL = fullURL + "?cursor=" + config.NextPage.String()
	}

	if config.Fields != nil {
		fields = config.Fields
	}

	res, err = c.get(ctx, fullURL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res, getTotalSize,
		func(node *ajson.Node) ([]map[string]interface{}, error) {
			return getRecords(node, config.ObjectName)
		},
		func(node *ajson.Node) (string, error) {
			return getNextRecordsURL(node, fullURL)
		},

		getMarshaledData,
		fields,
	)
}
