package closecrm

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Read retrieves data based on the provided read parameters.
// ref: https://developer.close.com/resources/leads/
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// Add _fields query parameters for filtering the response fields.
	url.WithQueryParamList("_fields", config.Fields.List())

	if !config.Since.IsZero() {
		// Filter response data according to the provided since data.
		url.WithQueryParam("date_updated", config.Since.Format(time.RFC3339))
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	return c.getAPIURL(config.ObjectName)
}
