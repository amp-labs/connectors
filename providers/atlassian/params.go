package atlassian

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

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Workspace.ValidateParams(),
		// Metadata parameter is optional.
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

// WithModule sets the Atlassian API module to use for the connector. It's required.
func WithModule(module paramsbuilder.APIModule) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, &ModuleEmpty)
	}
}

// WithMetadata sets authentication metadata expected by connector.
func WithMetadata(metadata map[string]string) Option {
	return func(params *parameters) {
		params.WithMetadata(metadata, nil)
	}
}
