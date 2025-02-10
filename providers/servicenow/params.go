package servicenow

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Module
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

// WithModule sets the ServiceNow API module to use for the connector. It's required.
// May need update. ServiceNow might have tenths of these modules.
func WithModule(module common.ModuleID) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, ModuleTable)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

func WithWorkspace(workspace string) Option {
	return func(params *parameters) {
		params.WithWorkspace(workspace)
	}
}
