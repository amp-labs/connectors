package freshdesk

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/parameters"
)

type Option = func(params *parametersInternal)

type parametersInternal struct {
	paramsbuilder.Client
	paramsbuilder.Workspace
}

func newParams(opts []Option) (*parameters.Connector, error) { // nolint:unused
	oldParams, err := paramsbuilder.Apply(parametersInternal{}, opts)
	if err != nil {
		return nil, err
	}

	return &parameters.Connector{
		AuthenticatedClient: oldParams.Client.Caller.Client,
		Workspace:           oldParams.Workspace.Name,
	}, nil
}

func (p parametersInternal) ValidateParams() error {
	return errors.Join(
		p.Client.ValidateParams(),
	)
}

func WithClient(
	ctx context.Context, client *http.Client,
	username string, password string,
	opts ...common.HeaderAuthClientOption,
) Option {
	return func(p *parametersInternal) {
		p.WithBasicClient(ctx, client, username, password, opts...)
	}
}

// WithWorkspace sets the freshdesk API instance to use for the connector. It's required.
func WithWorkspace(workspace string) Option {
	return func(params *parametersInternal) {
		params.WithWorkspace(workspace)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(p *parametersInternal) {
		p.WithAuthenticatedClient(client)
	}
}
