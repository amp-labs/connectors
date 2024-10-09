package atlassian

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
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

	rsp, err := c.Clients.JSON.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		getRecords,
		getNextRecords,
		common.GetMarshaledData,
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

	if !config.Since.IsZero() {
		// Read URL supports time scoping. common.ReadParams.Since is used to get relative time frame.
		// Here is an API example on how to request issues that were updated in the last 30 minutes.
		// search?jql=updated > "-30m"
		// The reason we use minutes is that it is the most granular API permits.
		diff := time.Since(config.Since)

		minutes := int64(diff.Minutes())
		if minutes > 0 {
			url.WithQueryParam("jql", fmt.Sprintf(`updated > "-%vm"`, minutes))
		}
	}

	return url, nil
}
