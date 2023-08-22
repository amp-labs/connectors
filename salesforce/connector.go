package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

const (
	apiVersion = "v58.0"
)

// Connector is a Salesforce connector.
type Connector struct {
	Domain  string
	BaseURL string
	client  common.HTTPClient
}

func (c *Connector) HTTPClient() common.HTTPClient { //nolint:ireturn
	return c.client
}

// NewConnector returns a new Salesforce connector.
func NewConnector(opts ...Option) (*Connector, error) {
	params := &sfParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error
	params, err = params.prepare()

	if err != nil {
		return nil, err
	}

	return &Connector{
		BaseURL: fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s", params.workspaceRef, apiVersion),
		Domain:  fmt.Sprintf("%s.my.salesforce.com", params.workspaceRef),
		client:  wrapClient(newHTTPClient(params)),
	}, nil
}
