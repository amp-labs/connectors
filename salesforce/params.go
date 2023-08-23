package salesforce

import (
	"net/http"

	"golang.org/x/oauth2"
)

type Option func(params *sfParams)

func WithClient(client *http.Client) Option {
	return func(params *sfParams) {
		params.client = client
	}
}

func WithOAuthToken(token *oauth2.Token) Option {
	return func(params *sfParams) {
		params.token = token
	}
}

func WithOAuthConfig(config *oauth2.Config) Option {
	return func(params *sfParams) {
		params.config = config
	}
}

func WithTokenSource(tokenSource oauth2.TokenSource) Option {
	return func(params *sfParams) {
		params.tokenSource = tokenSource
	}
}

func WithWorkspace(workspaceRef string) Option {
	return func(params *sfParams) {
		params.workspaceRef = workspaceRef
	}
}

type sfParams struct {
	client       *http.Client
	workspaceRef string
	token        *oauth2.Token
	config       *oauth2.Config
	tokenSource  oauth2.TokenSource
}

func (p *sfParams) prepare() (*sfParams, error) {
	if p.client == nil {
		p.client = http.DefaultClient
	}

	if len(p.workspaceRef) == 0 {
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
