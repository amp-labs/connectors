package docusign

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
)

const (
	restapiPrefix = "restapi"
	versionPrefix = "v2.1"
)

type Connector struct {
	BaseURL string
	Client  *common.JSONHTTPClient

	accountId string
}

func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	params, err := paramsbuilder.Apply(parameters{}, opts)
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	// Convert metadata map to model which knows how to do variable substitution.
	authMetadata := NewAuthMetadataVars(params.Metadata.Map)

	conn.accountId = authMetadata.AccountId

	// Read provider info
	providerInfo, err := providers.ReadInfo(providers.Docusign, authMetadata)
	if err != nil {
		return nil, err
	}

	// Set the base URL
	conn.setBaseURL(providerInfo.BaseURL)
	conn.Client.HTTPClient.ErrorHandler = interpreter.ErrorHandler{
		JSON: interpreter.NewFaultyResponder(errorFormats, nil),
	}.Handle

	return conn, nil
}

// Provider returns the connector provider.
func (c *Connector) Provider() providers.Provider {
	return providers.Docusign
}

func (c *Connector) setBaseURL(newURL string) {
	c.BaseURL = newURL
	c.Client.HTTPClient.Base = newURL
}
