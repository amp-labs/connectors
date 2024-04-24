package gong

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
)

type gongParams struct {
	client *common.JSONHTTPClient
	paramsbuilder.Workspace
	substitutions map[string]string
}

type Option func(params *gongParams)

func (p gongParams) FromOptions(opts ...Option) (*gongParams, error) {
	params := &p
	for _, opt := range opts {
		opt(params)
	}

	return params, params.ValidateParams()
}

func (p gongParams) ValidateParams() error {
	return errors.Join(
		p.Workspace.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
) Option {
	return func(params *gongParams) {
		oauthclient, err := common.NewOAuthHTTPClient(
			ctx, common.WithClient(client),
			common.WithOAuthConfig(config),
			common.WithOAuthToken(token),
		)
		if err != nil {
			panic(err)
		}

		WithAuthenticatedClient(oauthclient)(params)
	}
}

func WithWorkspace(workspaceRef string) Option {
	return func(params *gongParams) {
		params.WithWorkspace(workspaceRef)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *gongParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

func (params *gongParams) prepare() (*gongParams, error) {
	if params.client == nil {
		return nil, ErrMissingClient
	}

	return params, nil
}
