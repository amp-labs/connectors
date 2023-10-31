package hubspot

import (
	"context"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// search uses the POST /search endpoint to filter object records and return the result.
// This is used when Since is set. Otherwise, the Read endpoint is used.
// This endpoint paginates using paging.next.after which is to be used as an offset.
// Read more @ https://developers.hubspot.com/docs/api/crm/search
func (c *Connector) search(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var (
		data *ajson.Node
		err  error
	)

	// If filtering is not required, then we have to use the read endpoint.
	if !requiresFiltering(config) {
		return c.Read(ctx, config)
	}

	// If the next page is set, then we have to use the next page as the offset
	// in the filter body. As always, we attach the query values in the request.
	data, err = c.post(
		ctx,
		c.BaseURL+"/"+config.ObjectName+"/search"+"?"+makeQueryValues(config),
		makeFilterBody(config),
	)
	if err != nil {
		return nil, err
	}

	return parseResult(data, getNextRecordsAfter)
}

// makeFilterBody is specifically implemented for the Since filter currently.
func makeFilterBody(config common.ReadParams) map[string]any {
	filterBody := map[string]any{
		"filterGroups": []map[string]any{
			{
				"filters": []map[string]any{
					{
						"propertyName": "lastmodifieddate",
						"operator":     "GT",
						"value":        config.Since.Format(time.RFC3339),
					},
				},
			},
		},
	}

	if len(config.NextPage) > 0 {
		filterBody["after"] = config.NextPage
	}

	return filterBody
}
