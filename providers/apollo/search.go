package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// search uses POST method (GET for two objects[opportunities, tasks]) to read data
// from the Apollo io API. It has a limit of 50,000 records. It's recommended to filter
// the results so as to narrow down the results as much as possible.
func (c *Connector) search(ctx context.Context,
	url *urlbuilder.URL, config common.ReadParams,
) (*common.ReadResult, error) {
	// Check if searching the next page
	if len(config.NextPage) > 0 {
		url.WithQueryParam("page", config.NextPage.String())
	}

	json, err := c.Client.Post(ctx, url.String(), []byte{})
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		json,
		searchRecords(responseKey[config.ObjectName]),
		getNextRecords,
		common.GetMarshaledData,
		config.Fields,
	)
}
