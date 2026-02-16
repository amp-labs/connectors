package outreach

import (
	"context"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	includeQueryParam = "include"

	// defaultPageSize is the maximum page size supported by the Outreach API.
	// Setting page[size] explicitly forces the API to use cursor-based pagination (page[after] tokens)
	// instead of the deprecated offset-based pagination (page[offset]), which has a hard cap of 1000 records.
	defaultPageSize = "1000"
)

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and
// configuration parameters. It returns the nested Attributes values read results or an error
// if the operation fails.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	// Sets the query parameter `include` when the requests has request for associated objects.
	if len(config.AssociatedObjects) > 0 {
		url.WithQueryParam(includeQueryParam, strings.Join(config.AssociatedObjects, ","))
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	included, err := common.UnmarshalJSON[includedObjects](res)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		getOutreachDataMarshaller(config, included.Included, common.FlattenNestedFields(attributesKey)),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// Use page[size] to force cursor-based pagination instead of deprecated offset pagination.
	url.WithQueryParam("page[size]", defaultPageSize)

	// Outreach supports range filters: filter[updatedAt]=start..end
	// Note: Outreach only allows "inf" at the end of the range, not the start.
	// See: https://developers.outreach.io/api/making-requests/#filter-by-greater-than-or-equal-to-condition
	switch {
	case !config.Since.IsZero() && !config.Until.IsZero():
		url.WithQueryParam("filter[updatedAt]",
			config.Since.Format(time.DateOnly)+".."+config.Until.Format(time.DateOnly))
	case !config.Since.IsZero():
		url.WithQueryParam("filter[updatedAt]",
			config.Since.Format(time.DateOnly)+"..inf")
	case !config.Until.IsZero():
		url.WithQueryParam("filter[updatedAt]",
			"1970-01-01.."+config.Until.Format(time.DateOnly))
	}

	return url, nil
}
