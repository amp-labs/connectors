package zendesksupport

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parameters)

// parameters surface options by delegation.
type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Module
}

func newParams(opts []Option) (*common.Parameters, error) {
	oldParams, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(providers.ModuleZendeskTicketing),
	)
	if err != nil {
		return nil, err
	}

	return &common.Parameters{
		Module:              oldParams.Module.Selection.ID,
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
	}, nil
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
		params.WithOauthClient(ctx, client, config, token, opts...)
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

// WithModule sets the zendesk API module to use for the connector. It's required.
func WithModule(module common.ModuleID) Option {
	return func(params *parameters) {
		params.WithModule(module, SupportedModules, providers.ModuleZendeskTicketing)
	}
}
