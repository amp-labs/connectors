package hubspot

import (
	"context"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// Read reads data from Hubspot. If Since is set, it will use the
// Search endpoint instead to filter records, but it will be
// limited to a maximum of 10,000 records. This is a limit of the
// search endpoint. If Since is not set, it will use the read endpoint.
// In case Deleted objects won’t appear in any search results.
// Deleted objects can only be read by using this endpoint.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		rsp *common.JSONHTTPResponse
		err error
	)

	// If filtering is required, then we have to use the search endpoint.
	// The Search endpoint has a 10K record limit. In case this limit is reached,
	// the sorting allows the caller to continue in another call by offsetting
	// until the ID of the last record that was successfully fetched.
	if requiresFiltering(config) {
		searchParams := SearchParams{
			ObjectName: config.ObjectName,
			FilterGroups: []FilterGroup{
				{
					Filters: []Filter{
						BuildLastModifiedFilterGroup(&config),
						// Add more filters to AND them together
					},
					// Add more filter groups to OR them together
				},
			},
			SortBy: []SortBy{
				BuildSort(ObjectFieldHsObjectId, SortDirectionAsc),
			},
			NextPage: config.NextPage,
			Fields:   config.Fields,
		}

		return c.Search(ctx, searchParams)
	}

	if len(config.NextPage) > 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		rsp, err = c.Client.Get(ctx, config.NextPage.String())
	} else {
		// If NextPage is not set, then we're reading the first page of results.
		// We need to construct the query and then make the request.
		relativeURL := strings.Join([]string{"objects", config.ObjectName, "?" + makeQueryValues(config)}, "/")
		rsp, err = c.Client.Get(ctx, c.getURL(relativeURL))
	}

	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getTotalSize,
		getRecords,
		getNextRecordsURL,
		getMarshalledData,
		config.Fields,
	)
}

// makeQueryValues returns the query for the desired read operation.
func makeQueryValues(config common.ReadParams) string {
	queryValues := url.Values{}

	if len(config.Fields) != 0 {
		queryValues.Add("properties", strings.Join(config.Fields, ","))
	}

	if config.Deleted {
		queryValues.Add("archived", "true")
	}

	queryValues.Add("limit", DefaultPageSize)

	return queryValues.Encode()
}
