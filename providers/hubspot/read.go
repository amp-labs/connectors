package hubspot

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

// Read reads data from Hubspot. If Since is set, it will use the
// SearchLegacy endpoint instead to filter records, but it will be
// limited to a maximum of 10,000 records. This is a limit of the
// search endpoint. If Since is not set, it will use the read endpoint.
// In case Deleted objects won’t appear in any search results.
// Deleted objects can only be read by using this endpoint.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) { //nolint:funlen
	ctx = logging.With(ctx, "connector", "hubspot")

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if core.ObjectsWithoutPropertiesAPISupport.Has(config.ObjectName) {
		// Objects outside ObjectAPI have different endpoint while both are part of CRM module.
		// For instance Lists are fully returned only via Search endpoint.
		return c.searchCRM(ctx, searchCRMParams{
			SearchParams: SearchParams{
				ObjectName: config.ObjectName,
				NextPage:   config.NextPage,
				Fields:     config.Fields,
			},
		})
	}

	// If filtering is required, then we have to use the search endpoint.
	// The Search endpoint has a 10K record limit. In case this limit is reached,
	// the sorting allows the caller to continue in another call by offsetting
	// until the ID of the last record that was successfully fetched.
	filters := make(Filters, 0)
	if !config.Since.IsZero() {
		filters = append(filters, BuildLastModifiedFilterGroup(&config))
	}

	if !config.Until.IsZero() {
		filters = append(filters, BuildUntilTimestampFilterGroup(&config))
	}

	if len(filters) != 0 {
		searchParams := SearchParams{
			ObjectName: config.ObjectName,
			FilterGroups: []FilterGroup{{
				Filters: filters,
				// Add more filter groups to OR them together
			}},
			SortBy: []SortBy{
				BuildSort(ObjectFieldHsObjectId, SortDirectionAsc),
			},
			NextPage:          config.NextPage,
			Fields:            config.Fields,
			AssociatedObjects: config.AssociatedObjects,
		}

		return c.SearchLegacy(ctx, searchParams)
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsURL,
		associations.CreateDataMarshallerWithAssociations(
			ctx, c.crmAdapter.AssociationsFiller, config.ObjectName, config.AssociatedObjects),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (string, error) {
	if len(config.NextPage) != 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return config.NextPage.String(), nil
	}

	// If NextPage is not set, then we're reading the first page of results.
	// We need to construct the query and then make the request.
	// NB: The final slash is just to emulate prior behavior in earlier versions
	// of this code. If it turns out to be unnecessary, remove it.
	return c.getCRMObjectsReadURL(config)
}

// makeCRMObjectsQueryValues returns the query for the desired read operation.
func makeCRMObjectsQueryValues(config common.ReadParams) []string {
	var out []string

	fields := config.Fields.List()
	if len(fields) != 0 {
		out = append(out, "properties", strings.Join(fields, ","))
	}

	if config.Deleted {
		out = append(out, "archived", "true")
	}

	out = append(out, "limit", core.DefaultPageSize)

	return out
}
