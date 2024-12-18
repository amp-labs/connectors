package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and provided read parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		res *common.JSONHTTPResponse
		err error
	)

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(config.ObjectName, readOp)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(perPage, pageSize)

	// If NextPage is set, then we're reading the next page of results.
	if len(config.NextPage) > 0 {
		url.WithQueryParam(pageQuery, config.NextPage.String())
	}

	// If object uses POST searching, then we have to use the search endpoint, POST method.
	// The Search endpoint has a 50K record limit.
	switch {
	case in(config.ObjectName, readingSearchObjectsGET, readingListObjects):
		res, err = c.Client.Get(ctx, url.String())
		if err != nil {
			return nil, err
		}
	case in(config.ObjectName, readingSearchObjectsPOST):
		return c.Search(ctx, config)
	default:
		return nil, common.ErrObjectNotSupported
	}

	return common.ParseResult(res,
		recordsWrapperFunc(config.ObjectName),
		getNextRecords,
		common.GetMarshaledData,
		config.Fields,
	)
}
