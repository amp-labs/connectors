package docusign

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

type docusignParams struct {
	client *common.JSONHTTPClient
}

type Option func(params *docusignParams)

func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
) Option {
	return func(params *docusignParams) {
		oauthclient, err := common.NewOAuthHTTPClient(
			ctx, common.WithClient(client),
			common.WithOAuthConfig(config),
			common.WithOAuthToken(token),
		)
		if err != nil {
			panic(err)
		}

		WithAuthenticatedClient(oauthclient)(params)
	}
}

func WithAuthenticatedClient(client common.AuthenticatedHTTPClient) Option {
	return func(params *docusignParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

func (params *docusignParams) prepare() (*docusignParams, error) {
	if params.client == nil {
		return nil, ErrMissingClient
	}

	return params, nil
}
