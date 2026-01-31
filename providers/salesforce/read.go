package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

const defaultSOQLPageSize = 2000

// Read reads data from Salesforce. By default, it will read all rows (backfill). However, if Since is set,
// it will read only rows that have been updated since the specified time.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if c.isPardotModule() {
		return c.pardotAdapter.Read(ctx, config)
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
		core.GetRecords,
		core.GetNextRecordsURL,
		core.GetSalesforceDataMarshaller(config),
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
	// https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/resources_query.htm
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("q", makeSOQL(config).String())

	return url, nil
}

// makeSOQL returns the SOQL query for the desired read operation.
func makeSOQL(config common.ReadParams) *core.SOQLBuilder {
	fields := associations.FieldsForSelectQueryRead(&config)
	soql := (&core.SOQLBuilder{}).SelectFields(fields).From(config.ObjectName)
	addWhereClauses(soql, config)

	return soql
}

// addWhereClauses adds WHERE clauses to the SOQL query based on the config.
func addWhereClauses(soql *core.SOQLBuilder, config common.ReadParams) {
	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	if !config.Since.IsZero() {
		soql.Where("SystemModstamp > " + datautils.Time.FormatRFC3339inUTC(config.Since))
	}

	if !config.Until.IsZero() {
		soql.Where("SystemModstamp <= " + datautils.Time.FormatRFC3339inUTC(config.Until))
	}

	if config.Deleted {
		soql.Where("IsDeleted = true")
	}

	// TODO: When we support builder facing filters, we should escape the
	// filter string to avoid SOQL injection.
	if config.Filter != "" {
		soql.Where(config.Filter)
	}

	if config.PageSize > 0 {
		soql.Limit(int64(config.PageSize))
	}
}

func (c *Connector) DefaultPageSize() int {
	return defaultSOQLPageSize
}
