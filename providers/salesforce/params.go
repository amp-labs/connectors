package salesforce

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/salesforce/internal/pardot"
	"golang.org/x/oauth2"
)

// Option is a function which mutates the salesforce connector configuration.
type Option = func(params *parameters)

// parameters is the internal configuration for the salesforce connector.
type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
	paramsbuilder.Metadata
	paramsbuilder.Module

	timestampColumn string
	// provider allows the same Connector implementation to be reused under a
	// different provider name (e.g. salesforceJWT), which shares the underlying
	// Salesforce APIs but differs only in its authentication scheme. Defaults
	// to providers.Salesforce when unset.
	provider providers.Provider
}

func newParams(opts []Option) (*common.ConnectorParams, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parameters{}, opts,
		WithModule(providers.ModuleSalesforceCRM),
	)
	if err != nil {
		return nil, err
	}

	return &common.ConnectorParams{
		Module:              oldParams.Module.Selection.ID,
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
		Metadata:            oldParams.Map,
	}, nil
}

func (p parameters) ValidateParams() error {
	if isPardotModule(p.Module.Selection.ID) {
		// Check that business unit id is present.
		p.Metadata.WithMetadata(p.Metadata.Map, []string{
			pardot.MetadataKeyBusinessUnitID,
		})

		return errors.Join(
			p.Client.ValidateParams(),
			p.Metadata.ValidateParams(),
		)
	}

	return errors.Join(
		p.Client.ValidateParams(),
		p.Workspace.ValidateParams(),
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

func WithModule(module common.ModuleID) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, providers.ModuleSalesforceCRM)
	}
}

func WithMetadata(metadata map[string]string) Option {
	return func(params *parameters) {
		params.WithMetadata(metadata, nil)
	}
}

// WithTimestampColumn overrides the default "SystemModstamp" field used for
// incremental read queries. When set, the given field name is used in the
// Since/Until WHERE clauses instead of SystemModstamp.
func WithTimestampColumn(field string) Option {
	return func(params *parameters) {
		params.timestampColumn = field
	}
}

// WithProvider overrides the provider name this connector reports and uses
// when looking up ProviderInfo. It is intended for twin providers that reuse
// the same underlying Salesforce implementation but differ in auth scheme
// (e.g. salesforceJWT). When unset, the default is providers.Salesforce.
func WithProvider(provider providers.Provider) Option {
	return func(params *parameters) {
		params.provider = provider
	}
}
