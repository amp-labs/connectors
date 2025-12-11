package closecrm

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

/*
Response Schema:

{
  "data": [{...},{...}]
  cursor: "..."

}

*/

// searchEndpoint represents the search api endpoint.
// ref: https://developer.close.com/resources/advanced-filtering/
const searchEndpoint = "data/search"

// Search reads data through searching API. Supports advanced filtering using the filters field.
// The NextPage Token generated takes 30 seconds to expire.
//
// doc: https://developer.close.com/resources/advanced-filtering
func (c *Connector) Search(ctx context.Context, config SearchParams) (*common.ReadResult, error) {
	searchFilter, err := buildUpdatedDateFilter(config)
	if err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(searchEndpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(ctx, addTrailingSlashIfNeeded(url.String()), searchFilter)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.ExtractRecordsFromPath("data"),
		getNextRecordCursor,
		common.GetMarshaledData,
		datautils.NewStringSet(config.Fields...),
	)
}

func buildUpdatedDateFilter(params SearchParams) (Filter, error) {
	limit, err := strconv.Atoi(defaultPageSize)
	if err != nil {
		return Filter{}, err
	}

	flt := Filter{
		Query: Query{
			Type: "and",
			Queries: []map[string]any{
				{
					ObjectTypeQueryKey: params.ObjectName,
					TypeQueryKey:       "object_type",
				},
				{
					TypeQueryKey: "field_condition",
					FieldQueryKey: map[string]any{
						TypeQueryKey:          "regular_field",
						ObjectTypeQueryKey:    params.ObjectName,
						FieldNameTypeQueryKey: "date_updated",
					},
					ConditionQueryKey: map[string]any{
						OnOrAfterQueryKey: map[string]any{
							TypeQueryKey:  "fixed_local_date",
							ValueQueryKey: params.Since.Format(time.DateOnly),
							WhichQueryKey: "start",
						},
						TypeQueryKey: "moment_range",
					},
				},
			},
		},
		Fields: map[string][]string{
			params.ObjectName: params.Fields,
		},
		Cursor: nil,
		Limit:  limit,
	}

	if len(params.NextPage) > 0 {
		flt.Cursor = params.NextPage.String()
	}

	return flt, nil
}
