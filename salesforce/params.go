package salesforce

import (
	"net/http"

	"golang.org/x/oauth2"
)

// Option is a function which mutates the salesforce connector configuration.
type Option func(params *sfParams)

// WithClient sets the http client to use for the connector. Its usage is optional.
func WithClient(client *http.Client) Option {
	return func(params *sfParams) {
		params.client = client
	}
}

// WithOAuthToken sets the oauth token to use for the connector. It's required,
// unless a token source is provided.
func WithOAuthToken(token *oauth2.Token) Option {
	return func(params *sfParams) {
		params.token = token
	}
}

// WithOAuthConfig sets the oauth config to use for the connector. It's required,
// unless a token source is provided.
func WithOAuthConfig(config *oauth2.Config) Option {
	return func(params *sfParams) {
		params.config = config
	}
}

// WithTokenSource sets the oauth token source to use for the connector. Whenever
// the token expires, this will be called to refresh it.
func WithTokenSource(tokenSource oauth2.TokenSource) Option {
	return func(params *sfParams) {
		params.tokenSource = tokenSource
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
	client    *http.Client // optional
	subdomain string       // required

	// if tokenSource is provided, token and config are ignored.
	// if tokenSource is not provided, token and config are required.
	token       *oauth2.Token
	config      *oauth2.Config
	tokenSource oauth2.TokenSource
}

// prepare finalizes and validates the connector configuration, and returns an error if it's invalid.
func (p *sfParams) prepare() (*sfParams, error) {
	if p.client == nil {
		p.client = http.DefaultClient
	}

	if len(p.subdomain) == 0 {
		return nil, ErrMissingWorkspaceRef
	}

	if p.tokenSource == nil {
		if p.token == nil {
			return nil, ErrMissingRefreshToken
		}

		if p.config == nil {
			return nil, ErrMissingOauthConfig
		}
	}

	return p, nil
}
