package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

// get reads data from Salesforce. It handles retries and access token refreshes.
func (c *Connector) get(ctx context.Context, url string) (*common.JSONHTTPResponse, error) {
	node, err := c.Client.Get(ctx, url)
	if err != nil {
		return nil, c.HandleError(err)
	}

	// Success
	return node, nil
}

func (c *Connector) putCSV(ctx context.Context, url string, body []byte) ([]byte, error) {
	resBody, err := c.Client.PutCSV(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return resBody, nil
}

// post writes data to Salesforce. It handles retries and access token refreshes.
func (c *Connector) post(ctx context.Context, url string, body any) (*common.JSONHTTPResponse, error) {
	rsp, err := c.Client.Post(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return rsp, nil
}

// Patch writes data to Salesforce. It handles retries and access token refreshes.
func (c *Connector) patch(ctx context.Context, url string, body any) (*common.JSONHTTPResponse, error) {
	rsp, err := c.Client.Patch(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return rsp, nil
}
