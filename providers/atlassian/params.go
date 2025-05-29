package atlassian

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the connector configuration.
type Option = func(params *parametersInternal)

// parameters surface options by delegation.
type parametersInternal struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Module
	paramsbuilder.Metadata
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts,
		WithModule(common.ModuleRoot), // The module is resolved on behalf of the user if the option is missing.
	)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		Module:              oldParams.Module.Selection.ID,
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
		Metadata:            oldParams.Metadata.Map,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
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
	return func(params *parametersInternal) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parametersInternal) {
		params.WithAuthenticatedClient(client)
	}
}

func WithWorkspace(workspaceRef string) Option {
	return func(params *parametersInternal) {
		params.WithWorkspace(workspaceRef)
	}
}

// WithModule sets the Atlassian API module to use for the connector. It's required.
func WithModule(module common.ModuleID) Option {
	return func(params *parametersInternal) {
		params.WithModule(module, supportedModules, common.ModuleRoot)
	}
}

// WithMetadata sets authentication metadata expected by connector.
func WithMetadata(metadata map[string]string) Option {
	return func(params *parametersInternal) {
		params.WithMetadata(metadata, nil)
	}
}
