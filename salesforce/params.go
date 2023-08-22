package salesforce

import (
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

type Option func(params *sfParams)

func WithClient(client common.HTTPClient) Option {
	return func(params *sfParams) {
		params.client = client
	}
}

func WithToken(token *oauth2.Token) Option {
	return func(params *sfParams) {
		params.token = token
	}
}

func WithConfig(config *oauth2.Config) Option {
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
	client       common.HTTPClient
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
