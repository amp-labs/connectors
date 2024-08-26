package atlassian

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Delete removes Jira issue.
func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if len(config.RecordId) == 0 {
		return nil, common.ErrMissingRecordID
	}

	url, err := c.getJiraRestApiURL("issue")
	if err != nil {
		return nil, err
	}

	url.AddPath(config.RecordId)

	// 204 NoContent is expected
	_, err = c.Client.Delete(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return &common.DeleteResult{
		Success: true,
	}, nil
}
