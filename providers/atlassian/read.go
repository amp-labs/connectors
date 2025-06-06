package atlassian

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/atlassian/internal/jql"
)

// Read only returns a list of Jira Issues.
// You can provide the following values:
// * ObjectName - ignored.
// * NextPage - to get next page which may have no elements left.
// * Since - to scope the time frame, precision is in minutes.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecords,
		common.MakeMarshaledDataFunc(flattenRecord),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	url, err := c.getJiraRestApiURL("search")
	if err != nil {
		return nil, err
	}

	if len(config.NextPage) != 0 {
		url.WithQueryParam("startAt", config.NextPage.String())
	}

	jqlQuery := jql.New().
		SinceMinutes(config.Since).
		UntilMinutes(config.Until).
		String()

	if jqlQuery != "" {
		url.WithQueryParam("jql", jqlQuery)
	}

	return url, nil
}
