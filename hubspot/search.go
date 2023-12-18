package hubspot

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Search uses the POST /search endpoint to filter object records and return the result.
// This endpoint has a limit of 10,000 records. If the result has more than 10,000 records,
// the caller should employ sorting to paginate through the result on the client side.
// This endpoint paginates using paging.next.after which is to be used as an offset.
// Archived results do not appear in search results.
// Read more @ https://developers.hubspot.com/docs/api/crm/search
func (c *Connector) Search(ctx context.Context, config SearchParams) (*common.ReadResult, error) {
	var (
		data *ajson.Node
		err  error
	)

	data, err = c.post(ctx, c.BaseURL+"/objects/"+config.ObjectName+"/search", makeFilterBody(config))
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		data,
		getTotalSize,
		getRecords,
		getNextRecordsAfter,
		getMarshaledData,
		config.Fields,
	)
}

// BuildLastModifiedFilterGroup filters records modified since the given time.
func BuildLastModifiedFilterGroup(since time.Time) Filter {
	return Filter{
		FieldName: string(ObjectFieldLastModifiedDate),
		Operator:  FilterOperatorTypeGTE,
		Value:     since.Format(time.RFC3339),
	}
}

// BuildIdFilterGroup filters records greater than the given id.
func BuildIdFilterGroup(id string) Filter {
	return Filter{
		FieldName: string(ObjectFieldHsObjectId),
		Operator:  FilterOperatorTypeGT,
		Value:     id,
	}
}

// BuildSort builds a sort by clause for the given field and direction.
func BuildSort(field ObjectField, dir SortDirection) SortBy {
	return SortBy{
		PropertyName: string(field),
		Direction:    dir,
	}
}

func makeFilterBody(config SearchParams) map[string]any {
	filterBody := map[string]any{
		"limit": DefaultPageSize,
	}

	if config.FilterGroups != nil {
		filterBody["filterGroups"] = config.FilterGroups
	}

	if config.NextPage != "" {
		filterBody["after"] = config.NextPage
	}

	if config.SortBy != nil {
		filterBody["sorts"] = config.SortBy
	}

	if config.Fields != nil {
		filterBody["properties"] = config.Fields
	}

	return filterBody
}
