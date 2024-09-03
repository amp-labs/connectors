package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// search uses POST method to read data.It has a display limit of 50,000 records.
// It's recommended to filter the results so as to narrow down the results as much as possible.
// Most of the Filtering would need client's input so we don't exhaust calls by paging through all 50k records.
// Using this as is may lead to that issue.
func (c *Connector) Search(ctx context.Context, config common.ReadParams,
) (*common.ReadResult, error) {
	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// Adds searching path
	url.AddPath(searchingPath)

	// Check if searching the next page
	if len(config.NextPage) > 0 {
		url.WithQueryParam("page", config.NextPage.String())
	}

	// Add sorting & filtering criteria
	// Currently the default values, are what we needed
	// API sorts by the last activity or creation date timestamp.
	// So need to change the param details here.

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
