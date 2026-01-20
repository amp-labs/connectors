package hubspot

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// GetRecordCount returns the count of records for the given object and time range.
// It uses HubSpot's search endpoint with limit=0 to get the total count efficiently.
func (c *Connector) GetRecordCount(
	ctx context.Context,
	params *common.RecordCountParams,
) (*common.RecordCountResult, error) {
	// Build search URL
	url, err := c.getCRMObjectsSearchURL(SearchParams{
		ObjectName: params.ObjectName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	// Build filter body with minimal limit to reduce response size.
	// We only need the "total" field, not the actual records.
	filterBody := map[string]any{
		"limit": 1,
	}

	// Add time filters if provided
	filters := make(Filters, 0)

	if params.SinceTimestamp != nil && !params.SinceTimestamp.IsZero() {
		readParams := &common.ReadParams{
			ObjectName: params.ObjectName,
			Since:      *params.SinceTimestamp,
		}
		filters = append(filters, BuildLastModifiedFilterGroup(readParams))
	}

	if params.UntilTimestamp != nil && !params.UntilTimestamp.IsZero() {
		readParams := &common.ReadParams{
			ObjectName: params.ObjectName,
			Until:      *params.UntilTimestamp,
		}
		filters = append(filters, BuildUntilTimestampFilterGroup(readParams))
	}

	if len(filters) > 0 {
		filterBody["filterGroups"] = []FilterGroup{{
			Filters: filters,
		}}
	}

	// Execute the search request
	response, err := c.Client.Post(ctx, url, filterBody)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}

	// Parse the response to get total
	res, err := common.UnmarshalJSON[struct {
		Total int `json:"total"`
	}](response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return &common.RecordCountResult{
		Count: res.Total,
	}, nil
}
