package closecrm

import (
	"context"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
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
	input, err := buildUpdatedDateFilter(config)
	if err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(searchEndpoint)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(ctx, url.String(), input)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		common.GetRecordsUnderJSONPath("data"),
		searchNextRecords(),
		common.GetMarshaledData,
		config.Fields,
	)
}

func buildUpdatedDateFilter(config SearchParams) (Filters, error) {
	limit, err := strconv.Atoi(defaultPageSize)
	if err != nil {
		return Filters{}, err
	}

	flt := Filters{
		FilterQueries: FilterQueries{
			Type: "and",
			Queries: []map[string]any{
				{
					ObjectTypeQueryKey: config.ObjectName,
					TypeQueryKey:       "object_type",
				},
				{
					TypeQueryKey: "field_condition",
					FieldQueryKey: map[string]any{
						TypeQueryKey:          "regular_field",
						ObjectTypeQueryKey:    config.ObjectName,
						FieldNameTypeQueryKey: "date_updated",
					},
					ConditionQueryKey: map[string]any{
						OnOrAfterQueryKey: map[string]any{
							TypeQueryKey:  "fixed_local_date",
							ValueQueryKey: config.Since.Format(time.DateOnly),
							WhichQueryKey: "start",
						},
						TypeQueryKey: "moment_range",
					},
				},
			},
		},
		Fields: map[string][]string{
			config.ObjectName: config.Fields.List(),
		},
		Cursor: nil,
		Limit:  limit,
	}

	if len(config.NextPage) > 0 {
		flt.Cursor = config.NextPage.String()
	}

	return flt, nil
}
