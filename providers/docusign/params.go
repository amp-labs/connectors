package docusign

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
	"golang.org/x/oauth2"
)

type Option = func(params *parametersInternal)

type parametersInternal struct {
	paramsbuilder.Client
	paramsbuilder.Metadata
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Metadata:            oldParams.Metadata.Map,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
		p.Metadata.ValidateParams(),
	)
}

// WithClient sets the http client to use for the connector. Saves some boilerplate.
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

// WithMetadata sets authentication metadata expected by connector.
func WithMetadata(metadata map[string]string) Option {
	return func(params *parametersInternal) {
		params.WithMetadata(metadata, requiredMetadataFields)
	}
}
