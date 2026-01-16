package salesforce

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	crmcore "github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

// GetTotalRecords returns the total number of records for the given object and time range.
// It uses Salesforce's COUNT() SOQL function to get the count efficiently.
// This method only works for the CRM module, not for Pardot (Account Engagement).
func (c *Connector) GetTotalRecords(ctx context.Context, params *connectors.TotalRecordsParam) (int, error) {
	// Only support CRM module - Pardot doesn't use SOQL
	if c.isPardotModule() {
		return 0, fmt.Errorf("GetTotalRecords is not supported for Pardot (Account Engagement) module")
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
		return 0, fmt.Errorf("failed to build query URL: %w", err)
	}

	url.WithQueryParam("q", soql.String())

	// Execute the query
	response, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return 0, fmt.Errorf("failed to execute count query: %w", err)
	}

	// Parse the response to get totalSize
	type queryResponse struct {
		TotalSize int  `json:"totalSize"`
		Done      bool `json:"done"`
	}

	result, err := common.UnmarshalJSON[queryResponse](response)
	if err != nil {
		return 0, fmt.Errorf("failed to parse count query response: %w", err)
	}

	return result.TotalSize, nil
}
