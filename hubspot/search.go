package hubspot

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// Search uses the POST /search endpoint to filter object records and return the result.
// This is used when Since is set. Otherwise, the Read endpoint is used.
// This endpoint paginates using paging.next.after which is to be used as an offset.
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

	return parseResult(data, getNextRecordsAfter)
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

	return filterBody
}
