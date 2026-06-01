package servicenow

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Search performs a synchronous, filtered lookup against a ServiceNow table.
//
// ServiceNow's REST Table API exposes filtering via the sysparm_query parameter
// using an encoded-query string: `field1=value1^field2=value2` joins predicates
// with AND. Only equality is supported in common.SearchFilter today, which maps
// cleanly onto the `field=value` form.
func (c *Connector) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	url, err := c.constructSearchURL(params)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractRecordsFromPath("result"),
		getNextRecordsURL(resp),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) constructSearchURL(params *common.SearchParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	path, err := objectPath(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIPrefix, path)
	if err != nil {
		return nil, err
	}

	query, err := buildSysparmQuery(&params.Filter)
	if err != nil {
		return nil, err
	}

	if query != "" {
		url.WithQueryParam("sysparm_query", query)
	}

	if params.Fields != nil && !params.Fields.IsEmpty() {
		url.WithQueryParam("sysparm_fields", strings.Join(params.Fields.List(), ","))
	}

	if params.Limit > 0 {
		url.WithQueryParam("sysparm_limit", strconv.FormatInt(params.Limit, 10))
	}

	return url, nil
}

// buildSysparmQuery converts a SearchFilter into a ServiceNow encoded-query string.
// Multiple FieldFilters are AND-joined with `^`, matching the SearchFilter contract.
// Only FilterOperatorEQ is supported because that is the only operator in common.
func buildSysparmQuery(filter *common.SearchFilter) (string, error) {
	if filter == nil || len(filter.FieldFilters) == 0 {
		return "", nil
	}

	predicates := make([]string, 0, len(filter.FieldFilters))

	for _, ff := range filter.FieldFilters {
		if ff.Operator != common.FilterOperatorEQ {
			return "", fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, ff.Operator)
		}

		predicates = append(predicates, fmt.Sprintf("%s=%v", ff.FieldName, ff.Value))
	}

	return strings.Join(predicates, "^"), nil
}
