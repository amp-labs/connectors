package hubspot

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// Connector is a Hubspot connector.
type Connector struct {
	Module  string
	BaseURL string
	Client  *common.HTTPClient
}

// NewConnector returns a new Hubspot connector.
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

	params := &hubspotParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error
	params, err = params.prepare()

	params.client.Base = fmt.Sprintf("https://api.hubapi.com/%s", params.module)

	if err != nil {
		return nil, err
	}

	conn = &Connector{
		BaseURL: params.client.Base,
		Module:  params.module,
		Client:  params.client,
	}

	conn.Client.ErrorHandler = conn.interpretError

	return conn, nil
}
