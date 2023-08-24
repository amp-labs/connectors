package salesforce

import (
	"github.com/amp-labs/connectors/common"
)

// Option is a function which mutates the salesforce connector configuration.
type Option func(params *sfParams)

// WithClient sets the http client to use for the connector. Its usage is optional.
func WithClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *sfParams) {
		params.client = &common.JSONHTTPClient{
			Client:       client,
			ErrorHandler: common.InterpretError,
		}
	}
}

// WithSubdomain sets the salesforce subdomain to use for the connector. It's required.
func WithSubdomain(workspaceRef string) Option {
	return func(params *sfParams) {
		params.subdomain = workspaceRef
	}
}

// sfParams is the internal configuration for the salesforce connector.
type sfParams struct {
	client    *common.JSONHTTPClient // required
	subdomain string                 // required
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *sfParams) prepare() (*sfParams, error) {
	if p.client == nil {
		return nil, ErrMissingClient
	}

	if len(p.subdomain) == 0 {
		return nil, ErrMissingSubdomain
	}

	return p, nil
}
