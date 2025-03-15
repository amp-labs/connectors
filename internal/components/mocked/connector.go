package mocked

import (
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// Connector suitable for building different connector types for mock testing.
type Connector struct {
	// BaseURL must be set to the test server URL.
	BaseURL string
}

var _ connectors.Connector = &Connector{}

func (c Connector) String() string {
	return "mock_connector"
}

func (c Connector) JSONHTTPClient() *common.JSONHTTPClient {
	return &common.JSONHTTPClient{
		HTTPClient: &common.HTTPClient{
			Base:   c.BaseURL,
			Client: http.DefaultClient,
		},
	}
}

func (c Connector) HTTPClient() *common.HTTPClient {
	return c.JSONHTTPClient().HTTPClient
}

func (c Connector) Provider() providers.Provider {
	return "mock_test"
}
