package salesforce

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// BulkQuery launches async Query job for bulk reading.
//
// Design your SOQL query string you want to execute. Have a look at Salesforce documentation to know
// what clauses are not recommended for usage (under Request Body section).
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/query_create_job.htm
// To know more about SOQL syntax:
// https://developer.salesforce.com/docs/atlas.en-us.250.0.soql_sosl.meta/soql_sosl/sforce_api_calls_soql_select.htm
//
// After creation inspect newly launched Bulk Job via:
// * GetBulkQueryInfo
// * GetBulkQueryResults.
func (c *Connector) BulkQuery(
	ctx context.Context,
	query string,
	includeDeleted bool,
) (*GetJobInfoResult, error) {
	operation := "query"
	if includeDeleted {
		// Returns records that have been deleted because of a merge or delete.
		// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/query_create_job.htm
		operation = "queryAll"
	}

	jobBody := map[string]any{
		"operation": operation,
		"query":     query,
	}

	location, err := c.getRestApiURL("jobs/query")
	if err != nil {
		return nil, err
	}

	res, err := c.JSON.Post(ctx, location.String(), jobBody)
	if err != nil {
		return nil, fmt.Errorf("bulk query failed: %w", err)
	}

	return common.UnmarshalJSON[GetJobInfoResult](res)
}
