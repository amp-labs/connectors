package marketo

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

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
	config *clientcredentials.Config, token *oauth2.Token,
) Option {

	authClient := config.Client(ctx)

	return func(params *parameters) {
		params.WithAuthenticatedClient(authClient)
	}
}

// WithModule sets the marketo API module to use for the connector. It's required.
func WithModule(module paramsbuilder.APIModule) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, &ModuleEmpty)
	}
}

func WithWorkspace(workspace string) Option {
	return func(params *parameters) {
		params.WithWorkspace(workspace)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}
