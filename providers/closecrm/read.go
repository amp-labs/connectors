package closecrm

import (
	"context"
	"strconv"
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

	// The first page starts at offset zero.
	// The next page URL already carries its offset, which must not be overwritten.
	if len(config.NextPage) == 0 {
		url.WithQueryParam(skipQuery, "0")
		url.WithQueryParam(limitQuery, pageSize(config))
	}

	urlstring := addTrailingSlashIfNeeded(url.String())

	resp, err := c.Client.Get(ctx, urlstring)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.ExtractRecordsFromPath("data"),
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

// pageSize returns the requested page size, capped at maxPageSize.
// Close answers a _limit above the resource maximum with HTTP 400 instead of returning a smaller page,
// and the maximum is not documented per resource, so larger requests are reduced to the default.
// https://developer.close.com/topics/pagination/
func pageSize(config common.ReadParams) string {
	if config.PageSize <= 0 || config.PageSize > maxPageSize {
		return defaultPageSize
	}

	return strconv.Itoa(config.PageSize)
}
