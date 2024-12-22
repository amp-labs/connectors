package closecrm

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Read retrieves data based on the provided read parameters.
// ref: https://developer.close.com/resources/leads
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	// If Since is provided we Read data through the searchig API.
	// The searching API supports the incremental read using dates only.
	// The API has a limit of 10K records when paginating.
	// doc: https://developer.close.com/resources/advanced-filtering/
	if !config.Since.IsZero() && supportsFiltering(config.ObjectName) {
		return c.Search(ctx, SearchParams{
			ObjectName: config.ObjectName,
			Fields:     config.Fields.List(),
			Since:      config.Since,
			NextPage:   config.NextPage,
		})
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// Add initial query parameters for mutating the response fields.
	url.WithQueryParam("_fields", strings.Join(config.Fields.List(), ","))
	url.WithQueryParam(skipQuery, "0")
	url.WithQueryParam(limitQuery, defaultPageSize)

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.GetRecordsUnderJSONPath("data"),
		nextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	return c.getAPIURL(config.ObjectName)
}
