package paramsbuilder

import (
	"context"
	"errors"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"golang.org/x/oauth2"
)

var ErrMissingClient = errors.New("http client not set")

// Client params sets up authenticated proxy HTTP client.
// There are many types of authentication, where only one must be chosen. Ex: oauth2.
type Client struct {
	// Caller is an HTTP client that knows how to make authenticated requests.
	// It also knows how to handle authentication and API response errors.
	Caller *common.HTTPClient
}

func (p *Client) ValidateParams() error {
	// http client must be defined.
	if p.Caller == nil {
		return ErrMissingClient
	}

	// authentication client should be present.
	if p.Caller.Client == nil {
		return ErrMissingClient
	}

	return nil
}

// WithOauthClient option that sets up client that utilises Oauth2 authentication.
func (p *Client) WithOauthClient(
	ctx context.Context, client *http.Client,
	config *oauth2.Config, token *oauth2.Token,
	opts ...common.OAuthOption,
) {
	options := []common.OAuthOption{
		common.WithOAuthClient(client),
		common.WithOAuthConfig(config),
		common.WithOAuthToken(token),
	}

	oauthClient, err := common.NewOAuthHTTPClient(ctx, append(options, opts...)...)
	if err != nil {
		panic(err) // caught in NewConnector
	}

	p.WithAuthenticatedClient(oauthClient)
}

// WithBasicClient option that sets up client that utilises Basic (username, password) authentication.
func (p *Client) WithBasicClient(
	ctx context.Context, client *http.Client,
	user, pass string,
	opts ...common.HeaderAuthClientOption,
) {
	options := []common.HeaderAuthClientOption{
		common.WithHeaderClient(client),
	}

	basicClient, err := common.NewBasicAuthHTTPClient(ctx, user, pass, append(options, opts...)...)
	if err != nil {
		panic(err) // caught in NewConnector
	}

	p.WithAuthenticatedClient(basicClient)
}

// WithApiKeyHeaderClient option sets up client that utilises API Key authentication.
// Passed via Header.
func (p *Client) WithApiKeyHeaderClient(
	ctx context.Context, client *http.Client,
	headerName, headerValue string,
	opts ...common.HeaderAuthClientOption,
) {
	options := []common.HeaderAuthClientOption{
		common.WithHeaderClient(client),
	}

	apiKeyClient, err := common.NewApiKeyHeaderAuthHTTPClient(ctx,
		headerName, headerValue,
		append(options, opts...)...,
	)
	if err != nil {
		panic(err) // caught in NewConnector
	}

	p.WithAuthenticatedClient(apiKeyClient)
}

// WithApiKeyQueryParamClient option sets up client that utilises API Key authentication.
// Passed via Query Param.
func (p *Client) WithApiKeyQueryParamClient(
	ctx context.Context, client *http.Client,
	queryParamName, apiKey string,
	opts ...common.QueryParamAuthClientOption,
) {
	options := []common.QueryParamAuthClientOption{
		common.WithQueryParamClient(client),
	}

	apiKeyClient, err := common.NewApiKeyQueryParamAuthHTTPClient(ctx,
		queryParamName, apiKey,
		append(options, opts...)...,
	)
	if err != nil {
		panic(err) // caught in NewConnector
	}

	p.WithAuthenticatedClient(apiKeyClient)
}

// WithAuthenticatedClient sets up an HTTP client that uses your implementation of authentication.
func (p *Client) WithAuthenticatedClient(client common.AuthenticatedHTTPClient) {
	p.Caller = &common.HTTPClient{
		Client:       client,
		ErrorHandler: common.InterpretError,
	}
}
