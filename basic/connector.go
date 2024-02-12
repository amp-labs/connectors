package basic

import (
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

var (
	// ErrMissingClient is returned when a connector is created without a client.
	ErrMissingClient = errors.New("missing client")

	// ErrMissingBaseURL is returned when a connector is created without a base URL.
	ErrMissingBaseURL = errors.New("missing base URL")

	// ErrMissingProvider is returned when a connector is created without a provider.
	ErrMissingProvider = errors.New("missing provider")
)

// Connector is a Hubspot connector.
type Connector struct {
	ProviderInfo *providers.ProviderInfo `json:"providerInfo" mapstructure:"ProviderInfo"`
	Client       *common.JSONHTTPClient  `json:"client" mapstructure:"Client"`
	provider     providers.Provider
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

	params := &basicParams{}
	for _, opt := range opts {
		opt(params)
	}

	var err error
	params, err = params.prepare()
	if err != nil {
		return nil, err
	}

	conn = &Connector{
		provider: params.provider,
		Client:   params.client,
	}

	// Read provider info & replace catalog variables with given substitutions, if any
	providerInfo, err := providers.ReadConfig(conn.provider, &params.substitutions)
	if err != nil {
		return nil, err
	}

	conn.ProviderInfo = providerInfo
	conn.Client.HTTPClient.ErrorHandler = conn.interpretError

	// Set base URL
	conn.Client.HTTPClient.Base = conn.ProviderInfo.BaseURL

	return conn, nil
}
