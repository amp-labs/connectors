package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// get reads data from Salesforce. It handles retries and access token refreshes.
func (c *Connector) get(ctx context.Context, url string) (*ajson.Node, error) {
	node, err := c.Client.Get(ctx, url)
	if err != nil {
		return nil, c.HandleError(err)
	}

	// Success
	return node, nil
}

// get reads data from Salesforce. It handles retries and access token refreshes.
func (c *Connector) getCSV(ctx context.Context, url string) ([]byte, error) {
	body, err := c.Client.GetCSV(ctx, url, common.Header{
		Key:   "Accept",
		Value: "text/csv",
	})

	if err != nil {
		return nil, c.HandleError(err)
	}

	return body, nil

}
