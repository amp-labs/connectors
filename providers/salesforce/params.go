package salesforce

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the salesforce connector configuration.
type Option = func(params *parametersInternal)

// parameters is the internal configuration for the salesforce connector.
type parametersInternal struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Metadata
	paramsbuilder.Module
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts,
		WithModule(providers.ModuleSalesforceCRM),
	)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		Module:              oldParams.Module.Selection.ID,
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
	if isPardotModule(p.Module.Selection.ID) {
		return p.Client.ValidateParams()
	}

	return errors.Join(
		p.Client.ValidateParams(),
		p.Workspace.ValidateParams(),
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

func WithModule(module common.ModuleID) Option {
	return func(params *parametersInternal) {
		params.WithModule(module, supportedModules, providers.ModuleSalesforceCRM)
	}
}

func WithMetadata(metadata map[string]string) Option {
	return func(params *parametersInternal) {
		params.WithMetadata(metadata, nil)
	}
}
