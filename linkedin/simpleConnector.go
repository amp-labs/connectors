package linkedin

import (
	"github.com/amp-labs/connectors/common"
)

// Connector is a LinkedIn SimpleConnector.
type Connector struct {
	BaseURL string
	Client  *common.HTTPClient
}

// NewSimpleConnector returns a new Hubspot connector.
func NewSimpleConnector(opts ...Option) (conn *Connector, outErr error) {
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

	params := &linkedInParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error
	params, err = params.prepare()

	params.client.Base = "https://api.linkedin.com/v2"

	if err != nil {
		return nil, err
	}

	conn = &Connector{
		BaseURL: params.client.Base,
		Client:  params.client,
	}

	conn.Client.ErrorHandler = nil

	return conn, nil
}
