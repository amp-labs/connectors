package salesforce

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	crmcore "github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// GetRecordCount returns the count of records for the given object and time range.
// It uses Salesforce's COUNT() SOQL function to get the count efficiently.
//
// Example: counting Account records modified since a given time produces SOQL like:
//
//	SELECT COUNT() FROM Account WHERE SystemModstamp > 2024-01-15T00:00:00Z
//
// https://developer.salesforce.com/docs/atlas.en-us.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select_count.htm
func (c *Connector) GetRecordCount(
	ctx context.Context,
	params *common.RecordCountParams,
) (*common.RecordCountResult, error) {

	if c.isPardotModule() {
		return c.pardotAdapter.GetRecordCount(ctx, params)
	}

	// Build COUNT query
	soql := (&crmcore.SOQLBuilder{}).SelectCount().From(params.ObjectName)

	// Add WHERE clauses based on timestamps
	if params.SinceTimestamp != nil && !params.SinceTimestamp.IsZero() {
		soql.Where("SystemModstamp > " + datautils.Time.FormatRFC3339inUTC(*params.SinceTimestamp))
	}

	if params.UntilTimestamp != nil && !params.UntilTimestamp.IsZero() {
		soql.Where("SystemModstamp <= " + datautils.Time.FormatRFC3339inUTC(*params.UntilTimestamp))
	}

	// Build the query URL
	url, err := c.getRestApiURL("query")
	if err != nil {
		return nil, fmt.Errorf("failed to build query URL: %w", err)
	}

	url.WithQueryParam("q", soql.String())

	// Execute the query
	response, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to execute count query: %w", err)
	}

	// Parse the response to get totalSize
	res, err := common.UnmarshalJSON[struct {
		TotalSize int `json:"totalSize"`
	}](response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse count query response: %w", err)
	}

	return &common.RecordCountResult{
		Count: res.TotalSize,
	}, nil
}
