package zoho

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Read retrieves data based on the provided common.ReadParams configuration parameters.
// ref: https://www.zoho.com/crm/developer/docs/api/v6/get-records.html
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	switch c.moduleID {
	case providers.ZohoDeskV2:
		return c.read(ctx, config, nil)
	default:
		headers := constructHeaders(config)
		return c.read(ctx, config, headers)
	}
}

func (c *Connector) read(ctx context.Context, config common.ReadParams, headers []common.Header) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// If headers is nil, the variadic argument will be empty, making the call equivalent to Get(ctx, url.String())
	res, err := c.Client.Get(ctx, url.String(), headers...)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		extractRecordsFromPath(config.ObjectName),
		getNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func constructHeaders(config common.ReadParams) []common.Header {
	// Add the `If-Modified-Since` header if provided.
	// All Objects(or Modules in ZohoCRM terms) supports this.
	if !config.Since.IsZero() {
		return []common.Header{
			{
				Key:   "If-Modified-Since",
				Value: config.Since.Format(time.RFC3339),
			},
		}
	}

	return []common.Header{}
}
