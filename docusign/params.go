package docusign

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
)

type Option = func(params *parameters)

type parameters struct {
	paramsbuilder.Client
	paramsbuilder.Metadata
}

func (p parameters) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Metadata.ValidateParams(),
	)
}

// WithClient sets the http client to use for the connector. Saves some boilerplate.
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

// WithMetadata sets authentication metadata expected by connector.
func WithMetadata(metadata map[string]string) Option {
	return func(params *parameters) {
		params.WithMetadata(metadata, requiredMetadataFields)
	}
}
