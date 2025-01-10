package hubspot

import (
	"context"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
)

var (
	// maxHubspotReadQueryLength is the maximum length of a query that should be sent to Hubspot. This is around 2-3K,
	// but we add some buffer to be safe.
	maxHubspotReadQueryLength = 1800
)

// Read reads data from Hubspot. If Since is set, it will use the
// Search endpoint instead to filter records, but it will be
// limited to a maximum of 10,000 records. This is a limit of the
// search endpoint. If Since is not set, it will use the read endpoint.
// In case Deleted objects won’t appear in any search results.
// Deleted objects can only be read by using this endpoint.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) { //nolint:funlen
	ctx = logging.With(ctx, "connector", "hubspot")

	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	var (
		rsp *common.JSONHTTPResponse
		err error
	)

	// If filtering is required, then we have to use the search endpoint. This is also the case when the query is
	// too long. The Search endpoint has a 10K record limit. In case this limit is reached, the sorting allows the
	// caller to continue in another call by offsetting until the ID of the last record that was successfully fetched.
	if isIncrementalRead(config) || queryTooLong(config) {
		searchParams := buildSearchParams(config)

		return c.Search(ctx, searchParams)
	}

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		rsp, err = c.Client.Get(ctx, config.NextPage.String())
	} else {
		// If NextPage is not set, then we're reading the first page of results.
		// We need to construct the query and then make the request.
		// NB: The final slash is just to emulate prior behavior in earlier versions
		// of this code. If it turns out to be unnecessary, remove it.
		relativeURL := "objects/" + config.ObjectName + "/"

		u, urlErr := c.getURL(relativeURL, makeQueryValues(config)...)
		if urlErr != nil {
			return nil, urlErr
		}

		rsp, err = c.Client.Get(ctx, u)
	}

	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecordsURL,
		getMarshalledData,
		config.Fields,
	)
}

// makeQueryValues returns the query for the desired read operation.
func makeQueryValues(config common.ReadParams) []string {
	var out []string

	fields := config.Fields.List()
	if len(fields) != 0 {
		out = append(out, "properties", strings.Join(fields, ","))
	}

	if config.Deleted {
		out = append(out, "archived", "true")
	}

	out = append(out, "limit", DefaultPageSize)

	if len(config.AssociatedObjects) > 0 {
		out = append(out, "associations", strings.Join(config.AssociatedObjects, ","))
	}

	return out
}

func buildSearchParams(config common.ReadParams) SearchParams {
	searchParams := SearchParams{
		ObjectName: config.ObjectName,
		SortBy: []SortBy{
			BuildSort(ObjectFieldHsObjectId, SortDirectionAsc),
		},
		NextPage: config.NextPage,
		Fields:   config.Fields,
	}

	if isIncrementalRead(config) {
		searchParams.FilterGroups = []FilterGroup{
			{
				Filters: []Filter{
					BuildLastModifiedFilterGroup(&config),
					// Add more filters to AND them together
				},
				// Add more filter groups to OR them together
			},
		}
	}

	return searchParams
}

func queryTooLong(config common.ReadParams) bool {
	return len(strings.Join(config.Fields.List(), ",")) > maxHubspotReadQueryLength
}
