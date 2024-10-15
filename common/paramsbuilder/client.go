package paramsbuilder

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"golang.org/x/oauth2"
)

// Client params sets up authenticated proxy HTTP client.
// There are many types of authentication, where only one must be chosen. Ex: oauth2.
type Client struct {
	AuthClient
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
	provider providers.Provider, apiKey string,
	opts ...common.HeaderAuthClientOption,
) {
	info, err := providers.ReadInfo(provider)
	if err != nil {
		panic(err)
	}

	headerName, headerValue, err := info.GetApiKeyHeader(apiKey)
	if err != nil {
		panic(err)
	}

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
	provider providers.Provider, apiKey string,
	opts ...common.QueryParamAuthClientOption,
) {
	info, err := providers.ReadInfo(provider)
	if err != nil {
		panic(err)
	}

	queryParamName, err := info.GetApiKeyQueryParamName()
	if err != nil {
		panic(err)
	}

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
