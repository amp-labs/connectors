package atlassian

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Delete removes Jira issue.
func (c *Connector) Delete(ctx context.Context, config common.DeleteParams) (*common.DeleteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	url, err := c.getModuleURL("issue")
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
