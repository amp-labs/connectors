package kit

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/kit/metadata"
	"github.com/spyzhov/ajson"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(rsp,
		makeGetRecords(responseFieldName),
		makeNextRecordsURL(url),
		common.MakeMarshaledDataFunc(flattenCustomFields),
		config.Fields,
	)
}

// makeGetRecords creates a NodeRecordsFunc that extracts records from the API response
// using the specified field name. The field name corresponds to the array field in
// Kit's response that contains the list of records.
func makeGetRecords(responseFieldName string) common.NodeRecordsFunc {
	return func(node *ajson.Node) ([]*ajson.Node, error) {
		return jsonquery.New(node).ArrayOptional(responseFieldName)
	}
}

func (c *Connector) buildURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		// Next page.
		return constructURL(config.NextPage.String())
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("per_page", strconv.Itoa(DefaultPageSize))

	return url, nil
}
