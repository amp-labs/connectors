package outreach

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"golang.org/x/oauth2"
)

type outreachParams struct {
	client *common.JSONHTTPClient
}

type Option func(params *outreachParams)

func WithClient(ctx context.Context, client *http.Client, config *oauth2.Config, token *oauth2.Token,
) Option {
	return func(params *outreachParams) {
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
	return func(params *outreachParams) {
		params.client = &common.JSONHTTPClient{
			HTTPClient: &common.HTTPClient{
				Client:       client,
				ErrorHandler: common.InterpretError,
			},
		}
	}
}

func (params *outreachParams) prepare() (*outreachParams, error) {
	if params.client == nil {
		return nil, paramsbuilder.ErrMissingClient
	}

	return params, nil
}
