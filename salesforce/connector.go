package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

const (
	apiVersion = "v59.0"
)

// Connector is a Salesforce connector.
type Connector struct {
	Domain  string
	BaseURL string
	Client  *common.JSONHTTPClient
}

func (c *Connector) APIVersion() string {
	return apiVersion
}

// NewConnector returns a new Salesforce connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer func() {
		if re := recover(); re != nil {
			tmp, ok := re.(error)
			if !ok {
				panic(re)
			}

			outErr = tmp
			conn = nil
		}
	}()

	params := &sfParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error
	params, err = params.prepare()

	params.client.Base = fmt.Sprintf("https://%s.my.salesforce.com", params.subdomain)

	if err != nil {
		return nil, err
	}

	conn = &Connector{
		BaseURL: fmt.Sprintf("https://%s.my.salesforce.com/services/data/%s", params.subdomain, apiVersion),
		Domain:  fmt.Sprintf("%s.my.salesforce.com", params.subdomain),
		Client:  params.client,
	}

	conn.Client.ErrorHandler = conn.interpretError

	return conn, nil
}
