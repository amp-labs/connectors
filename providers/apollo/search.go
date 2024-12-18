package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// search uses POST method to read data.It has a display limit of 50,000 records.
// It's recommended to filter the results so as to narrow down the results as much as possible.
// Most of the Filtering would need client's input so we don't exhaust calls by paging through all 50k records.
// Using this as is may lead to that issue.
func (c *Connector) Search(ctx context.Context, config common.ReadParams, url *urlbuilder.URL,
) (*common.ReadResult, error) {
	resp, err := c.Client.Post(ctx, url.String(), []byte{})
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		searchRecords(config.ObjectName),
		getNextRecords,
		common.GetMarshaledData,
		config.Fields,
	)
}
