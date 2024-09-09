package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Read reads data from Salesforce. By default, it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return c.getDomainURL(config.NextPage.String())
	}

	// If NextPage is not set, then we're reading the first page of results.
	// We need to construct the SOQL query and then make the request.
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("q", makeSOQL(config).String())

	return url, nil
}

// makeSOQL returns the SOQL query for the desired read operation.
func makeSOQL(config common.ReadParams) *soqlBuilder {
	soql := (&soqlBuilder{}).SelectFields(config.Fields).From(config.ObjectName)

	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	if !config.Since.IsZero() {
		soql.Where("SystemModstamp > " + config.Since.Format("2006-01-02T15:04:05Z"))
	}

	if config.Deleted {
		soql.Where("IsDeleted = true")
	}

	// TODO: When we support builder facing filters, we should escape the
	// filter string to avoid SOQL injection.
	if config.Filter != "" {
		soql.Where(config.Filter)
	}

	return soql
}
