package zohocrm

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Read retrieves data based on the provided common.ReadParams configuration parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	modHeaders := constructHeader(config, url)

	res, err := c.Client.Get(ctx, url.String(), modHeaders...)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		common.GetRecordsUnderJSONPath("data"),
		getNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func constructHeader(config common.ReadParams, url *urlbuilder.URL) []common.Header {
	// Retrieve the since from the url.
	since, ok := url.GetFirstQueryParam("since")
	if !ok {
		return []common.Header{}
	}

	// Add the `If-Modified-Since` header if provided.
	// All Objects(or Modules in ZohoCRM terms) supports this.
	if !config.Since.IsZero() {
		modHeader := common.Header{
			Key:   "If-Modified-Since",
			Value: since,
		}

		return []common.Header{modHeader}
	}

	return []common.Header{}
}
