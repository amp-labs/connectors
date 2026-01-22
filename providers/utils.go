package providers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template

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

func (c CustomCatalog) catalog() (*CatalogWrapper, error) { // nolint:funcorder
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

func (i *ProviderInfo) SubstituteWith(vars catalogreplacer.CatalogVariables) error {
	// Take care of default metadata values.
	if i.Metadata != nil {
		for _, metadataInput := range i.Metadata.Input {
			if metadataInput.DefaultValue != "" {
				vars.AddDefaults(catalogreplacer.CustomCatalogVariable{Plan: catalogreplacer.SubstitutionPlan{
					From: metadataInput.Name,
					To:   metadataInput.DefaultValue,
				}})
			}
		}

		// To prevent OAuthConnect from erroring out due to missing PostAuthentication variables,
		// we add a default value of "" for each PostAuthentication variable.
		// Since no Connection exists yet, there won't be any PostAuthentication variables.
		for _, postAuthVar := range i.Metadata.PostAuthentication {
			vars.AddDefaults(catalogreplacer.CustomCatalogVariable{Plan: catalogreplacer.SubstitutionPlan{
				From: postAuthVar.Name,
				To:   "",
			}})
		}
	}

	return catalogreplacer.NewCatalogSubstitutionRegistry(vars).Apply(i)
}

func (i *ProviderInfo) GetOption(key string) (string, bool) {
	if i.ProviderOpts == nil {
		return "", false
	}

	val, ok := i.ProviderOpts[key]

	return val, ok
}

// ReadModuleInfo finds information about the module.
// If module is not found fallbacks to the default.
func (i *ProviderInfo) ReadModuleInfo(moduleID common.ModuleID) *ModuleInfo {
	// Empty value fallback to the default value defined in ProviderInfo.
	if moduleID == "" {
		moduleID = i.defaultModuleOrRoot()
	}

	// RootModule is inferred using the core values of ProviderInfo.
	// On the surface those connectors are module agnostic.
	rootModule := ModuleInfo{
		BaseURL:     i.BaseURL,
		DisplayName: i.DisplayName,
		Support:     i.Support,
	}

	// No modules exist. Fallback to the one and only RootModule.
	if !i.hasModules() {
		if moduleID != common.ModuleRoot {
			// TODO the catalog should be checked almost at the "compile time".
			// TODO There should be tests to ensure integrity. When anything is changed it should do consistency check.
			slog.Warn("provider doesn't have modules while a module was requested",
				"provider", i.DisplayName, "module", moduleID)
		}

		// Requesting root when no modules exist is allowed.
		return &rootModule
	}

	// Root module is providerInfo derived module.
	if moduleID == common.ModuleRoot {
		return &rootModule
	}

	// Find module.
	module, ok := (*i.Modules)[moduleID] // nolint:varnamelen
	if ok {
		return &module
	}

	// Invalid module requested.
	slog.Warn("module info is missing for a module",
		"provider", i.DisplayName, "module", moduleID)

	// Use fallback module to handle invalid module.
	fallbackModule := i.defaultModuleOrRoot()

	if fallbackModule == common.ModuleRoot {
		return &rootModule
	}

	module, ok = (*i.Modules)[fallbackModule]
	if !ok {
		slog.Warn("finding fallback module failed",
			"provider", i.DisplayName, "module", fallbackModule)

		return &rootModule
	}

	return &module
}

func (i *ProviderInfo) hasModules() bool {
	return i.Modules != nil && len(*i.Modules) != 0
}

func (i *ProviderInfo) defaultModuleOrRoot() common.ModuleID {
	if i.DefaultModule == "" {
		if i.hasModules() {
			slog.Warn("defaulting to root while provider supports multiple modulus",
				"provider", i.DisplayName)
		}

		return common.ModuleRoot
	}

	return i.DefaultModule
}

// UnauthorizedHandler is a function that is called when an unauthorized response is received.
// The handler can be used to refresh the token or to perform other actions. The client is
// included so you can make additional requests if needed, but be careful not to create an
// infinite loop (hint, use the request's context to attach a counter to avoid this possibility).
//
// The proper semantics of this function can be read as: "I received an unauthorized response,
// and I want to do something about it, and then I want to return a modified response to the
// original caller."
//
// The most common planned use case is to refresh the token and then retry the request and return
// the non-401 response to the original caller.
type UnauthorizedHandler func(client common.AuthenticatedHTTPClient, event *UnauthorizedEvent) (*http.Response, error)

// IsUnauthorizedDecider is a function called to determine if a response is unauthorized.
type IsUnauthorizedDecider func(rsp *http.Response) (bool, error)

// UnauthorizedEvent is the event that is triggered when an unauthorized response (http 401) is received.
type UnauthorizedEvent struct {
	Provider    *ProviderInfo
	Headers     []common.Header     // Only certain providers will set this, depending on the auth type
	QueryParams []common.QueryParam // Only certain providers will set this, depending on the auth type
	OAuthToken  *oauth2.Token       // Only certain providers will set this, depending on the auth type
	Request     *http.Request
	Response    *http.Response
}

// BasicParams is the parameters to create a basic auth client.
type BasicParams struct {
	User    string
	Pass    string
	Options []common.HeaderAuthClientOption
}

// ApiKeyParams is the parameters to create an api key client.
type ApiKeyParams struct {
	Key           string
	HeaderOptions []common.HeaderAuthClientOption
	QueryOptions  []common.QueryParamAuthClientOption
}

// OAuth2AuthCodeParams is the parameters to create an OAuth2 auth code client.
type OAuth2AuthCodeParams struct {
	Config  *oauth2.Config
	Token   *oauth2.Token
	Options []common.OAuthOption
}

// OAuth2ClientCredentialsParams is the parameters to create an OAuth2 client credentials client.
type OAuth2ClientCredentialsParams struct {
	Config  *clientcredentials.Config
	Options []common.OAuthOption
}

type CustomAuthParams struct {
	Values  map[string]string
	Options []common.CustomAuthClientOption
}

// NewClientParams is the parameters to create a new HTTP client.
type NewClientParams struct {
	// Debug will enable debug mode for the client.
	Debug bool

	// Client is the http client to use for the client. If
	// the value is nil, the default http client will be used.
	Client *http.Client

	// OnUnauthorized is the handler to call when the client receives an
	// unauthorized response.
	OnUnauthorized UnauthorizedHandler

	// IsUnauthorized is the function to call to determine if the response is unauthorized.
	// If not set, it will default to the dumb logic of checking for 401 status codes.
	IsUnauthorized IsUnauthorizedDecider

	// BasicCreds is the basic auth credentials to use for the client.
	// If the provider uses basic auth, this field must be set.
	BasicCreds *BasicParams

	// OAuth2ClientCreds is the client credentials to use for the client.
	// If the provider uses client credentials, this field must be set.
	OAuth2ClientCreds *OAuth2ClientCredentialsParams

	// OAuth2AuthCodeCreds is the auth code credentials to use for the client.
	// If the provider uses auth code, this field must be set.
	OAuth2AuthCodeCreds *OAuth2AuthCodeParams

	// ApiKeyCreds is the api key to use for the client. If the provider uses
	// api-key auth, this field must be set.
	ApiKeyCreds *ApiKeyParams

	// CustomCreds is the custom auth credentials to use for the client. If the provider uses
	// custom auth, this field must be set.
	CustomCreds *CustomAuthParams
}

// NewClient will create a new authenticated client based on the provider's auth type.
func (i *ProviderInfo) NewClient(ctx context.Context, params *NewClientParams) (common.AuthenticatedHTTPClient, error) { //nolint:lll,cyclop,ireturn,funlen
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
			return createOAuth2AuthCodeHTTPClient(
				ctx, params.Client, params.Debug, params.OnUnauthorized, params.IsUnauthorized, i, params.OAuth2AuthCodeCreds)
		case ClientCredentials:
			return createOAuth2ClientCredentialsHTTPClient(
				ctx, params.Client, params.Debug, params.OnUnauthorized, params.IsUnauthorized, i, params.OAuth2ClientCreds)
		case Password:
			return createOAuth2PasswordHTTPClient(
				ctx, params.Client, params.Debug, params.OnUnauthorized, params.IsUnauthorized, i, params.OAuth2AuthCodeCreds)
		default:
			return nil, fmt.Errorf("%w: unsupported grant type %q", ErrClient, i.Oauth2Opts.GrantType)
		}
	case Basic:
		if params.BasicCreds == nil {
			return nil, fmt.Errorf("%w: %s", ErrClient, "basic credentials not found")
		}

		return createBasicAuthHTTPClient(
			ctx, params.Client, params.Debug, params.OnUnauthorized, params.IsUnauthorized, i,
			params.BasicCreds.User, params.BasicCreds.Pass, params.BasicCreds.Options)
	case ApiKey:
		if i.ApiKeyOpts == nil {
			return nil, fmt.Errorf("%w: api key options not found", ErrClient)
		}

		if params.ApiKeyCreds == nil {
			return nil, fmt.Errorf("%w: api key credentials not found", ErrClient)
		}

		if len(params.ApiKeyCreds.Key) == 0 {
			return nil, fmt.Errorf("%w: api key not given", ErrClient)
		}

		return createApiKeyHTTPClient(ctx, params.Client, params.Debug, params.OnUnauthorized,
			params.IsUnauthorized, i, params.ApiKeyCreds)
	case Custom:
		if i.CustomOpts == nil {
			return nil, fmt.Errorf("%w: custom options not found", ErrClient)
		}

		if params.CustomCreds == nil {
			return nil, fmt.Errorf("%w: custom credentials not found", ErrClient)
		}

		return createCustomHTTPClient(ctx, params.Client, params.Debug, params.OnUnauthorized,
			params.IsUnauthorized, i, params.CustomCreds)
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
	unauth UnauthorizedHandler,
	isUnauth IsUnauthorizedDecider,
	info *ProviderInfo,
	user string,
	pass string,
	options []common.HeaderAuthClientOption,
) (common.AuthenticatedHTTPClient, error) {
	opts := []common.HeaderAuthClientOption{
		common.WithHeaderClient(getClient(client)),
	}

	if dbg {
		opts = append(opts, common.WithHeaderDebug(common.PrintRequestAndResponse))
	}

	var authClient common.AuthenticatedHTTPClient

	if isUnauth != nil {
		opts = append(opts, common.WithHeaderIsUnauthorizedHandler(isUnauth))
	}

	if unauth != nil {
		opts = append(opts,
			common.WithHeaderUnauthorizedHandler(
				func(hdrs []common.Header, req *http.Request, rsp *http.Response) (*http.Response, error) {
					return unauth(authClient, &UnauthorizedEvent{
						Provider: info,
						Headers:  hdrs,
						Request:  req,
						Response: rsp,
					})
				}))
	}

	if len(options) > 0 {
		opts = append(opts, options...)
	}

	var err error

	authClient, err = common.NewBasicAuthHTTPClient(ctx, user, pass, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create basic auth client: %w", ErrClient, err)
	}

	return authClient, nil
}

func createOAuth2AuthCodeHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	unauth UnauthorizedHandler,
	isUnauth IsUnauthorizedDecider,
	info *ProviderInfo,
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

	var oauthClient common.AuthenticatedHTTPClient

	if isUnauth != nil {
		options = append(options, common.WithOAuthIsUnauthorizedHandler(isUnauth))
	}

	if unauth != nil {
		options = append(options,
			common.WithOAuthUnauthorizedHandler(
				func(token *oauth2.Token, req *http.Request, rsp *http.Response) (*http.Response, error) {
					return unauth(oauthClient, &UnauthorizedEvent{
						Provider:   info,
						OAuthToken: token,
						Request:    req,
						Response:   rsp,
					})
				}))
	}

	if header := CreateOauth2TokenHeaderAttachment(info); header != nil {
		options = append(options, common.WithTokenHeaderAttachment(header))
	}

	options = append(options, cfg.Options...)

	var err error

	oauthClient, err = common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create oauth2 client: %w", ErrClient, err)
	}

	return oauthClient, nil
}

func createOAuth2ClientCredentialsHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	unauth UnauthorizedHandler,
	isUnauth IsUnauthorizedDecider,
	info *ProviderInfo,
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

	var oauthClient common.AuthenticatedHTTPClient

	if isUnauth != nil {
		options = append(options, common.WithOAuthIsUnauthorizedHandler(isUnauth))
	}

	if unauth != nil {
		options = append(options,
			common.WithOAuthUnauthorizedHandler(
				func(token *oauth2.Token, req *http.Request, rsp *http.Response) (*http.Response, error) {
					return unauth(oauthClient, &UnauthorizedEvent{
						Provider:   info,
						OAuthToken: token,
						Request:    req,
						Response:   rsp,
					})
				}))
	}

	if header := CreateOauth2TokenHeaderAttachment(info); header != nil {
		options = append(options, common.WithTokenHeaderAttachment(header))
	}

	options = append(options, cfg.Options...)

	var err error

	oauthClient, err = common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create oauth2 client: %w", ErrClient, err)
	}

	return oauthClient, nil
}

func createOAuth2PasswordHTTPClient(
	ctx context.Context,
	client *http.Client,
	dbg bool,
	unauth UnauthorizedHandler,
	isUnauth IsUnauthorizedDecider,
	info *ProviderInfo,
	cfg *OAuth2AuthCodeParams,
) (common.AuthenticatedHTTPClient, error) {
	// Refresh method works the same as with auth code method.
	// Relies on access and refresh tokens created by Oauth2 password method.
	return createOAuth2AuthCodeHTTPClient(ctx, client, dbg, unauth, isUnauth, info, cfg)
}

func createCustomHTTPClient(ctx context.Context, //nolint:funlen,cyclop
	client *http.Client,
	dbg bool,
	unauth UnauthorizedHandler,
	isUnauth IsUnauthorizedDecider,
	info *ProviderInfo,
	cfg *CustomAuthParams,
) (common.AuthenticatedHTTPClient, error) {
	// Make sure that all the inputs are provided in the config values.
	for _, input := range info.CustomOpts.Inputs {
		val, ok := cfg.Values[input.Name]
		if !ok || val == "" {
			return nil, fmt.Errorf("%w: missing value for custom client input %q",
				ErrClient, input.Name)
		}
	}

	// Get the static headers
	headers, err := getCustomHeaders(info, cfg)
	if err != nil {
		return nil, err
	}

	// Get the static query parameters
	queryParams, err := getCustomParams(info, cfg)
	if err != nil {
		return nil, err
	}

	var opts []common.CustomAuthClientOption

	if len(headers) > 0 {
		opts = append(opts, common.WithCustomHeaders(headers...))
	}

	if len(queryParams) > 0 {
		opts = append(opts, common.WithCustomQueryParams(queryParams...))
	}

	opts = append(opts, common.WithCustomClient(getClient(client)))

	if dbg {
		opts = append(opts, common.WithCustomDebug(common.PrintRequestAndResponse))
	}

	var customClient common.AuthenticatedHTTPClient

	if isUnauth != nil {
		opts = append(opts, common.WithCustomIsUnauthorizedHandler(isUnauth))
	}

	if unauth != nil {
		opts = append(opts,
			common.WithCustomUnauthorizedHandler(
				func(
					hdrs []common.Header,
					params []common.QueryParam,
					req *http.Request,
					rsp *http.Response,
				) (*http.Response, error) {
					return unauth(customClient, &UnauthorizedEvent{
						Headers:     hdrs,
						QueryParams: params,
						Provider:    info,
						Request:     req,
						Response:    rsp,
					})
				}))
	}

	if len(cfg.Options) > 0 {
		opts = append(opts, cfg.Options...)
	}

	customClient, err = common.NewCustomAuthHTTPClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create custom auth client: %w", ErrClient, err)
	}

	return customClient, nil
}

func getCustomParams(info *ProviderInfo, cfg *CustomAuthParams) (common.QueryParams, error) {
	if len(info.CustomOpts.QueryParams) == 0 {
		return nil, nil
	}

	params := make([]common.QueryParam, 0, len(info.CustomOpts.QueryParams))

	for _, param := range info.CustomOpts.QueryParams {
		value, err := evalTemplate(param.ValueTemplate, cfg.Values)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to evaluate query param template for param %q: %w",
				ErrClient, param.Name, err)
		}

		params = append(params, common.QueryParam{
			Key:   param.Name,
			Value: value,
		})
	}

	return params, nil
}

func getCustomHeaders(info *ProviderInfo, cfg *CustomAuthParams) (common.Headers, error) {
	if len(info.CustomOpts.Headers) == 0 {
		return nil, nil
	}

	headers := make([]common.Header, 0, len(info.CustomOpts.Headers))

	for _, hdr := range info.CustomOpts.Headers {
		value, err := evalTemplate(hdr.ValueTemplate, cfg.Values)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to evaluate header template for header %q: %w",
				ErrClient, hdr.Name, err)
		}

		headers = append(headers, common.Header{
			Key:   hdr.Name,
			Value: value,
		})
	}

	return headers, nil
}

func evalTemplate(input string, vars map[string]string) (string, error) {
	tmpl, err := template.New("-").Option("missingkey=error").Parse(input)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, vars); err != nil {
		return "", err
	}

	return sb.String(), nil
}

func createApiKeyHTTPClient( //nolint:ireturn,cyclop,funlen
	ctx context.Context,
	client *http.Client,
	dbg bool,
	unauth UnauthorizedHandler,
	isUnauth IsUnauthorizedDecider,
	info *ProviderInfo,
	cfg *ApiKeyParams,
) (common.AuthenticatedHTTPClient, error) {
	apiKey := cfg.Key

	switch info.ApiKeyOpts.AttachmentType {
	case Header:
		if info.ApiKeyOpts.Header.ValuePrefix != "" {
			apiKey = info.ApiKeyOpts.Header.ValuePrefix + apiKey
		}

		opts := []common.HeaderAuthClientOption{
			common.WithHeaderClient(getClient(client)),
		}

		if dbg {
			opts = append(opts, common.WithHeaderDebug(common.PrintRequestAndResponse))
		}

		var authClient common.AuthenticatedHTTPClient

		if isUnauth != nil {
			opts = append(opts, common.WithHeaderIsUnauthorizedHandler(isUnauth))
		}

		if unauth != nil {
			opts = append(opts,
				common.WithHeaderUnauthorizedHandler(
					func(hdrs []common.Header, req *http.Request, rsp *http.Response) (*http.Response, error) {
						return unauth(authClient, &UnauthorizedEvent{
							Provider: info,
							Headers:  hdrs,
							Request:  req,
							Response: rsp,
						})
					}))
		}

		if len(cfg.HeaderOptions) > 0 {
			opts = append(opts, cfg.HeaderOptions...)
		}

		var err error

		authClient, err = common.NewApiKeyHeaderAuthHTTPClient(ctx, info.ApiKeyOpts.Header.Name, apiKey, opts...)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create api key client: %w", ErrClient, err)
		}

		return authClient, nil
	case Query:
		opts := []common.QueryParamAuthClientOption{
			common.WithQueryParamClient(getClient(client)),
		}

		if dbg {
			opts = append(opts, common.WithQueryParamDebug(common.PrintRequestAndResponse))
		}

		var authClient common.AuthenticatedHTTPClient

		if isUnauth != nil {
			opts = append(opts, common.WithQueryParamIsUnauthorizedHandler(isUnauth))
		}

		if unauth != nil {
			opts = append(opts,
				common.WithQueryParamUnauthorizedHandler(
					func(params []common.QueryParam, req *http.Request, rsp *http.Response) (*http.Response, error) {
						return unauth(authClient, &UnauthorizedEvent{
							Provider:    info,
							QueryParams: params,
							Request:     req,
							Response:    rsp,
						})
					}))
		}

		if len(cfg.QueryOptions) > 0 {
			opts = append(opts, cfg.QueryOptions...)
		}

		var err error

		authClient, err = common.NewApiKeyQueryParamAuthHTTPClient(ctx, info.ApiKeyOpts.Query.Name, apiKey, opts...)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create api key client: %w", ErrClient, err)
		}

		return authClient, nil
	default:
		return nil, fmt.Errorf("%w: unsupported api key type %q", ErrClient, info.ApiKeyOpts.AttachmentType)
	}
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

func (i *ProviderInfo) RequiresWorkspace() bool {
	if i.Metadata == nil || i.Metadata.Input == nil {
		return false
	}

	for _, input := range i.Metadata.Input {
		if input.Name == "workspace" {
			// When default value is present then workspace is not required.
			// Missing the default makes workspace required.
			return input.DefaultValue == ""
		}
	}

	return false
}

// CreateOauth2TokenHeaderAttachment builds and returns a custom token header configuration
// for OAuth2 authentication, if the provider defines one.
//
// By default, OAuth2 tokens are sent using the standard
//
//	Authorization: Bearer <token>
//
// header. Some providers override this behavior and require a custom header instead
// (for example, Shopify uses: X-Shopify-Access-Token: <token>).
//
// If the provider does not specify a custom header configuration, this function returns
// (nil, false). Otherwise, it returns the configured TokenHeaderAttachment and true.
func CreateOauth2TokenHeaderAttachment(info *ProviderInfo) *common.TokenHeaderAttachment {
	if info.Oauth2Opts == nil ||
		info.Oauth2Opts.AccessTokenOpts == nil ||
		info.Oauth2Opts.AccessTokenOpts.Header == nil {
		return nil
	}

	header := info.Oauth2Opts.AccessTokenOpts.Header

	return &common.TokenHeaderAttachment{
		Name:   header.Name,
		Prefix: header.ValuePrefix,
	}
}
