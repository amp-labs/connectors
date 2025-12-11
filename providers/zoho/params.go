package zoho

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// Option is a functional parameter for configuring Zoho connector.
type Option = func(params *parameters)

// parameters holds the configuration for the Zoho connector.
type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Module

	// location is the Zoho data center location (e.g., "us", "eu", "in", "au", "jp", "ca").
	location string
	// domains contains the region-specific API endpoint URLs for the given location.
	domains *LocationDomains
}

// ValidateParams validates the connector parameters.
func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

// WithClient configures the connector with an OAuth2 authenticated client.
func WithClient(ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token, opts ...common.OAuthOption,
) Option {
	return func(params *parameters) {
		params.WithOauthClient(ctx, client, config, token, opts...)
	}
}

// WithAuthenticatedClient configures the connector with a pre-authenticated HTTP client.
func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *parameters) {
		params.WithAuthenticatedClient(client)
	}
}

// WithModule specifies the Zoho module (e.g., CRM, Books, Desk) to connect to.
func WithModule(module common.ModuleID) Option {
	return func(params *parameters) {
		params.WithModule(module, supportedModules, providers.ModuleZohoCRM)
	}
}

// WithLocation sets the Zoho data center location (e.g., "us", "eu", "in", "au", "jp", "ca").
// This determines which regional API endpoints will be used.
func WithLocation(location string) Option {
	return func(params *parameters) {
		params.location = location
	}
}

// WithDomains sets custom API endpoint URLs for a specific Zoho region.
// This allows overriding the default domain mappings for advanced use cases.
func WithDomains(domains *LocationDomains) Option {
	return func(params *parameters) {
		params.domains = domains
	}
}
