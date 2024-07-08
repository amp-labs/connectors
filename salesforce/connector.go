package salesforce

import (
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

const (
	providerOptionRestApiURL = "restApiUrl"
	providerOptionDomain     = "domain"
)

const (
	apiVersion    = "59.0"
	versionPrefix = "v"
)

// Connector is a Salesforce connector.
type Connector struct {
	Domain  string
	BaseURL string
	Client  *common.JSONHTTPClient
}

func APIVersion() string {
	return versionPrefix + apiVersion
}

func APIVersionSOAP() string {
	return apiVersion
}

// NewConnector returns a new Salesforce connector.
func NewConnector(opts ...Option) (conn *Connector, outErr error) {
	defer common.PanicRecovery(func(cause error) {
		outErr = cause
		conn = nil
	})

	params, err := parameters{}.FromOptions(opts...)
	if err != nil {
		return nil, err
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadInfo(providers.Salesforce, &params.Workspace)
	if err != nil {
		return nil, err
	}

	restApi, ok := providerInfo.GetOption(providerOptionRestApiURL)
	if !ok {
		return nil, fmt.Errorf("restApiUrl not set: %w", providers.ErrProviderOptionNotFound)
	}

	domain, ok := providerInfo.GetOption(providerOptionDomain)
	if !ok {
		return nil, fmt.Errorf("domain not set: %w", providers.ErrProviderOptionNotFound)
	}

	conn = &Connector{
		BaseURL: restApi,
		Domain:  domain,
		Client: &common.JSONHTTPClient{
			HTTPClient: params.Client.Caller,
		},
	}

	conn.Client.HTTPClient.Base = providerInfo.BaseURL
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError
	conn.Client.ErrorPostProcessor.Process = handleError

	return conn, nil
}
