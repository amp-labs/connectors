package generic

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

var (
	// ErrMissingClient is returned when a connector is created without a client.
	ErrMissingClient = errors.New("missing client")

	// ErrMissingProvider is returned when a connector is created without a provider.
	ErrMissingProvider = errors.New("missing provider")
)

type Option = func(*parameters)

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Metadata

	provider providers.Provider
}

func (p parameters) ValidateParams() error {
	if p.provider == "" {
		return ErrMissingProvider
	}

	// workspace is optional

	return errors.Join(
		p.Client.ValidateParams(),
	)
}

// WithClient sets the http client to use for the connector.
func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parameters) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

// WithAuthenticatedClient sets the http client to use for the connector. Its usage is optional.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

// WithWorkspace sets workspace which is used as substitution for URL templates.
func WithWorkspace(workspaceRef string) Option {
	return func(params *parameters) {
		params.WithWorkspace(workspaceRef)
	}
}

// WithMetadata sets authentication metadata expected by connector.
func WithMetadata(metadata map[string]string) Option {
	return func(params *parameters) {
		params.WithMetadata(metadata, nil)
	}
}

func (p parameters) GetCatalogVars() []catalogreplacer.CatalogVariable {
	variables := []catalogreplacer.CatalogVariable{
		&p.Workspace,
	}

	for _, v := range p.Metadata.GetCatalogVars() {
		variables = append(variables, v)
	}

	return variables
}
