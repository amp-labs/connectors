package salesforce

import (
	"context"
)

func (c *Connector) putCSV(ctx context.Context, url string, body []byte) ([]byte, error) {
	resBody, err := c.Client.PutCSV(ctx, url, body)
	if err != nil {
		return nil, c.HandleError(err)
	}

	return resBody, nil
}
