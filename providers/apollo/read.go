package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and provided read parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	url, err := c.getURL(config)
	if err != nil {
		return nil, err
	}

	// If the given object uses search endpoint for Reading,
	// checks for the  method and makes the call.
	if usesSearching(config.ObjectName) {
		url = url.AddPath(searchingPath)
		if in(config.ObjectName, postSearchObjects) {
			return c.search(ctx, url, config)
		}
	}

	// Objects that uses listing to read data by default
	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		recordsWrapperFunc(config.ObjectName),
		getNextRecords,
		common.GetMarshaledData,
		config.Fields,
	)
}
