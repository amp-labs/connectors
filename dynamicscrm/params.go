package dynamicscrm

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
)

const (
	// DefaultPageSize is number of elements per page.
	DefaultPageSize = 100
)

// Option is a function which mutates the connector configuration.
type Option func(params *parameters)

// parameters MS CRM supports auth client, workspace, etc. by delegation.
type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Module
}

func (p parameters) FromOptions(opts ...Option) (*parameters, error) {
	params := &p
	for _, opt := range opts {
		opt(params)
	}

	return params, params.ValidateParams()
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Workspace.ValidateParams(),
		p.Module.ValidateParams(),
	)
}

func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parameters) {
		params.WithClient(ctx, client, config, token, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

func WithWorkspace(workspaceRef string) Option {
	return func(params *parameters) {
		params.WithWorkspace(workspaceRef)
	}
}

func WithModule(module paramsbuilder.APIModule) Option {
	return func(params *parameters) {
		params.WithModule(module)
	}
}
