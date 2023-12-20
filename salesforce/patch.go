package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// Patch writes data to Salesforce. It handles retries and access token refreshes.
func (c *Connector) patch(ctx context.Context, url string, body any) (*common.JSONHTTPResponse, error) {
	rsp, err := c.Client.Patch(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return rsp, nil
}
