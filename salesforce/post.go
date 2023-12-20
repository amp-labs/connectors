package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// post writes data to Salesforce. It handles retries and access token refreshes.
func (c *Connector) post(ctx context.Context, url string, body any) (*common.JSONHTTPResponse, error) {
	rsp, err := c.Client.Post(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return rsp, nil
}
