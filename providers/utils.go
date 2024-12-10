// nolint:ireturn
package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/go-playground/validator"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	ErrCatalogNotFound                = errors.New("catalog not found")
	ErrProviderNotFound               = errors.New("provider not found")
	ErrClient                         = errors.New("client creation failed")
	ErrRetrievingHeaderApiKeyName     = errors.New("provider information missing header name for API Key")
	ErrRetrievingQueryParamApiKeyName = errors.New("provider information missing query parameter name for API Key")
)

type CatalogOption func(params *catalogParams)

type catalogParams struct {
	catalog *CatalogWrapper
}

// WithCatalog is an option that can be used to override the default catalog.
func WithCatalog(c *CatalogWrapper) CatalogOption {
	return func(params *catalogParams) {
		params.catalog = c
	}
}

type CustomCatalog struct {
	custom *CatalogWrapper
}

// NewCustomCatalog allows to apply modifiers on the base catalog, to tweak its content.
// Just like the default catalog it supports reading data, resolves variable substitutions.
func NewCustomCatalog(opts ...CatalogOption) CustomCatalog {
	params := &catalogParams{catalog: getCatalog()}

	for _, opt := range opts {
		opt(params)
	}

	return CustomCatalog{custom: params.catalog}
}

func (c CustomCatalog) catalog() (*CatalogWrapper, error) {
	if c.custom == nil {
		// Null catalog was probably set via options.
		// This is not allowed.
		return nil, ErrCatalogNotFound
	}

	return c.custom, nil
}

// ReadCatalog is used to get the catalog.
func ReadCatalog(opts ...CatalogOption) (*CatalogWrapper, error) {
	return NewCustomCatalog(opts...).ReadCatalog()
}

func (c CustomCatalog) ReadCatalog() (*CatalogWrapper, error) {
	catalogInstance, err := c.catalog()
	if err != nil {
		return nil, err
	}

	catalogCopy, err := goutils.Clone[*CatalogWrapper](catalogInstance)
	if err != nil {
		return nil, err
	}

	// Validate the provider configuration
	v := validator.New()
	for provider, providerInfo := range catalogCopy.Catalog {
		if err := v.Struct(providerInfo); err != nil {
			return nil, err
		}

		providerInfo.Name = provider
	}

	return catalogCopy, nil
}

// SetInfo sets the information for a specific provider in the catalog.
// This is useful to enable experimental providers or to override the default
// provider information. This is primarily used to initialize the provider catalog.
// Generally speaking, once the provider catalog is initialized, it should not be modified.
// That having been said, there are some use cases where it is useful to override the
// provider information, such as when testing new configurations. This function is not
// thread-safe and should be called before the provider catalog is read.
func SetInfo(provider Provider, info ProviderInfo) {
	if catalog == nil {
		catalog = make(CatalogType)
	}

	info.Name = provider

	catalog[provider] = info
}

// ReadInfo reads the information from the catalog for specific provider. It also performs string substitution
// on the values in the config that are surrounded by {{}}, if vars are provided.
// The catalog variable will be applied such that `{{.VAR_NAME}}` string will be replaced with `VAR_VALUE`.
func ReadInfo(provider Provider, vars ...catalogreplacer.CatalogVariable) (*ProviderInfo, error) {
	return NewCustomCatalog().ReadInfo(provider, vars...)
}

func (c CustomCatalog) ReadInfo(provider Provider, vars ...catalogreplacer.CatalogVariable) (*ProviderInfo, error) {
	catalogInstance, err := c.catalog()
	if err != nil {
		return nil, err
	}

	pInfo, ok := catalogInstance.Catalog[provider]
	if !ok {
		return nil, ErrProviderNotFound
	}

	// No substitution needed
	if len(vars) == 0 {
		return &pInfo, nil
	}

	// Clone before modifying
	providerInfo, err := goutils.Clone[ProviderInfo](pInfo)
	if err != nil {
		return nil, err
	}

	providerInfo.Name = provider

	// Validate the provider configuration
	v := validator.New()
	if err := v.Struct(providerInfo); err != nil {
		return nil, err
	}

	// Apply substitutions to the provider configuration values which contain variables in the form of {{var}}.
	if err := providerInfo.SubstituteWith(vars); err != nil {
		return nil, err
	}

	return &providerInfo, nil
}

func (i *ProviderInfo) SubstituteWith(vars []catalogreplacer.CatalogVariable) error {
	return catalogreplacer.NewCatalogSubstitutionRegistry(vars).Apply(i)
}

func (i *ProviderInfo) GetOption(key string) (string, bool) {
	if i.ProviderOpts == nil {
		return "", false
	}

	val, ok := i.ProviderOpts[key]

	return val, ok
}

// BasicParams is the parameters to create a basic auth client.
type BasicParams struct {
	User string
	Pass string
}

// OAuth2AuthCodeParams is the parameters to create an OAuth2 auth code client.
type OAuth2AuthCodeParams struct {
	Config  *oauth2.Config
	Token   *oauth2.Token
	Options []common.OAuthOption
}

type OAuth2ClientCredentialsParams struct {
	Config  *clientcredentials.Config
	Options []common.OAuthOption
}

// NewClientParams is the parameters to create a new HTTP client.
type NewClientParams struct {
	// Debug will enable debug mode for the client.
	Debug bool

	// Client is the http client to use for the client. If
	// the value is nil, the default http client will be used.
	Client *http.Client

	// BasicCreds is the basic auth credentials to use for the client.
	// If the provider uses basic auth, this field must be set.
	BasicCreds *BasicParams

	// OAuth2ClientCreds is the client credentials to use for the client.
	// If the provider uses client credentials, this field must be set.
	OAuth2ClientCreds *OAuth2ClientCredentialsParams

	// OAuth2AuthCodeCreds is the auth code credentials to use for the client.
	// If the provider uses auth code, this field must be set.
	OAuth2AuthCodeCreds *OAuth2AuthCodeParams

	// ApiKey is the api key to use for the client. If the provider uses api-key
	// auth, this field must be set.
	ApiKey string
}

// NewClient will create a new authenticated client based on the provider's auth type.
func (i *ProviderInfo) NewClient(ctx context.Context, params *NewClientParams) (common.AuthenticatedHTTPClient, error) { //nolint:lll,cyclop,ireturn
	if params == nil {
		params = &NewClientParams{}
	}

	switch i.AuthType {
	case None:
		return createUnauthenticatedClient(ctx, params.Client, params.Debug)
	case Oauth2:
		if i.Oauth2Opts == nil {
			return nil, fmt.Errorf("%w: %s", ErrClient, "oauth2 options not found")
		}

		switch i.Oauth2Opts.GrantType {
		case AuthorizationCodePKCE:
			fallthrough
		case AuthorizationCode:
			return createOAuth2AuthCodeHTTPClient(ctx, params.Client, params.Debug, params.OAuth2AuthCodeCreds)
		case ClientCredentials:
			return createOAuth2ClientCredentialsHTTPClient(ctx, params.Client, params.Debug, params.OAuth2ClientCreds)
		case Password:
			return createOAuth2PasswordHTTPClient(ctx, params.Client, params.Debug, params.OAuth2AuthCodeCreds)
		default:
			return nil, fmt.Errorf("%w: unsupported grant type %q", ErrClient, i.Oauth2Opts.GrantType)
		}
	case Basic:
		if params.BasicCreds == nil {
			return nil, fmt.Errorf("%w: %s", ErrClient, "basic credentials not found")
		}

		return createBasicAuthHTTPClient(ctx, params.Client, params.Debug, params.BasicCreds.User, params.BasicCreds.Pass)
	case ApiKey:
		if i.ApiKeyOpts == nil {
			return nil, fmt.Errorf("%w: api key options not found", ErrClient)
		}

		if len(params.ApiKey) == 0 {
			return nil, fmt.Errorf("%w: api key not given", ErrClient)
		}

		return createApiKeyHTTPClient(ctx, params.Client, params.Debug, i, params.ApiKey)
	case Jwt:
		// We shouldn't hit this case, because no providerInfo has auth type set to JWT yet.
		fallthrough
	default:
		return nil, fmt.Errorf("%w: unsupported auth type %q", ErrClient, i.AuthType)
	}
}

func getClient(client *http.Client) *http.Client {
	if client == nil {
		return http.DefaultClient
	}

	return client
}

func createUnauthenticatedClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
) (common.AuthenticatedHTTPClient, error) {
	opts := []common.HeaderAuthClientOption{
		common.WithHeaderClient(getClient(client)),
	}

	if dbg {
		opts = append(opts, common.WithHeaderDebug(common.PrintRequestAndResponse))
	}

	return common.NewHeaderAuthHTTPClient(ctx, opts...)
}

func createBasicAuthHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	user string,
	pass string,
) (common.AuthenticatedHTTPClient, error) {
	opts := []common.HeaderAuthClientOption{
		common.WithHeaderClient(getClient(client)),
	}

	if dbg {
		opts = append(opts, common.WithHeaderDebug(common.PrintRequestAndResponse))
	}

	c, err := common.NewBasicAuthHTTPClient(ctx, user, pass, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create basic auth client: %w", ErrClient, err)
	}

	return c, nil
}

func createOAuth2AuthCodeHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	cfg *OAuth2AuthCodeParams,
) (common.AuthenticatedHTTPClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("%w: oauth2 config not provided", ErrClient)
	}

	options := []common.OAuthOption{
		common.WithOAuthClient(getClient(client)),
		common.WithOAuthConfig(cfg.Config),
		common.WithOAuthToken(cfg.Token),
	}

	if dbg {
		options = append(options, common.WithOAuthDebug(common.PrintRequestAndResponse))
	}

	options = append(options, cfg.Options...)

	oauthClient, err := common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create oauth2 client: %w", ErrClient, err)
	}

	return oauthClient, nil
}

func createOAuth2ClientCredentialsHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	cfg *OAuth2ClientCredentialsParams,
) (common.AuthenticatedHTTPClient, error) {
	if _, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); !ok {
		if client != nil {
			ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
		}
	}

	options := []common.OAuthOption{
		common.WithOAuthClient(getClient(client)),
		common.WithTokenSource(cfg.Config.TokenSource(ctx)),
	}

	if dbg {
		options = append(options, common.WithOAuthDebug(common.PrintRequestAndResponse))
	}

	options = append(options, cfg.Options...)

	oauthClient, err := common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create oauth2 client: %w", ErrClient, err)
	}

	return oauthClient, nil
}

func createOAuth2PasswordHTTPClient(
	ctx context.Context,
	client *http.Client,
	dbg bool,
	cfg *OAuth2AuthCodeParams,
) (common.AuthenticatedHTTPClient, error) {
	// Refresh method works the same as with auth code method.
	// Relies on access and refresh tokens created by Oauth2 password method.
	return createOAuth2AuthCodeHTTPClient(ctx, client, dbg, cfg)
}

func createApiKeyHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	info *ProviderInfo,
	apiKey string,
) (common.AuthenticatedHTTPClient, error) {
	if info.ApiKeyOpts.AttachmentType == Header { //nolint:nestif
		if info.ApiKeyOpts.Header.ValuePrefix != "" {
			apiKey = info.ApiKeyOpts.Header.ValuePrefix + apiKey
		}

		opts := []common.HeaderAuthClientOption{
			common.WithHeaderClient(getClient(client)),
		}

		if dbg {
			opts = append(opts, common.WithHeaderDebug(common.PrintRequestAndResponse))
		}

		c, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, info.ApiKeyOpts.Header.Name, apiKey, opts...)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create api key client: %w", ErrClient, err)
		}

		return c, nil
	} else if info.ApiKeyOpts.AttachmentType == Query {
		opts := []common.QueryParamAuthClientOption{
			common.WithQueryParamClient(getClient(client)),
		}

		if dbg {
			opts = append(opts, common.WithQueryParamDebug(common.PrintRequestAndResponse))
		}

		c, err := common.NewApiKeyQueryParamAuthHTTPClient(ctx, info.ApiKeyOpts.Query.Name, apiKey, opts...)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create api key client: %w", ErrClient, err)
		}

		return c, nil
	}

	return nil, fmt.Errorf("%w: unsupported api key type %q", ErrClient, info.ApiKeyOpts.AttachmentType)
}

func (i *ProviderInfo) GetApiKeyQueryParamName() (string, error) {
	if i.ApiKeyOpts == nil || i.ApiKeyOpts.Query == nil || len(i.ApiKeyOpts.Query.Name) == 0 {
		return "", ErrRetrievingQueryParamApiKeyName
	}

	return i.ApiKeyOpts.Query.Name, nil
}

func (i *ProviderInfo) GetApiKeyHeader(apiKey string) (string, string, error) {
	if i.ApiKeyOpts == nil || i.ApiKeyOpts.Header == nil || len(i.ApiKeyOpts.Header.Name) == 0 {
		return "", "", ErrRetrievingHeaderApiKeyName
	}

	headerName := i.ApiKeyOpts.Header.Name

	headerValue := apiKey
	if i.ApiKeyOpts.Header.ValuePrefix != "" {
		// The prefix is non-empty, which means it is required.
		headerValue = i.ApiKeyOpts.Header.ValuePrefix + apiKey
	}

	return headerName, headerValue, nil
}

// Override can be used to override the base URL of the provider, and could be
// used for other fields in the future.
func (i *ProviderInfo) Override(override *ProviderInfo) *ProviderInfo {
	if i == nil {
		return &ProviderInfo{}
	}

	// Return the original if the override is nil.
	if override == nil {
		return i
	}

	// Only allow overriding the base URL for now.
	if override.BaseURL != "" {
		i.BaseURL = override.BaseURL
	}

	return i
}

func ReadInfoMap(provider Provider, vars map[string]string) (*ProviderInfo, error) {
	return NewCustomCatalog().ReadInfoMap(provider, vars)
}

func (c CustomCatalog) ReadInfoMap(provider Provider, vars map[string]string) (*ProviderInfo, error) {
	// TODO: Reads providerInfo and substitutes placeholders with a map
	return c.ReadInfo(provider)
}
