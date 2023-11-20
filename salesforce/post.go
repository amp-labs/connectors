package salesforce

import (
	"context"

	"github.com/spyzhov/ajson"
)

// post writes data to Salesforce. It handles retries and access token refreshes.
func (c *Connector) post(ctx context.Context, url string, body any) (*ajson.Node, error) {
	node, err := c.Client.Post(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return node, nil
}
