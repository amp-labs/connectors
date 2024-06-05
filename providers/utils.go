// nolint:ireturn
package providers

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"text/template" // nosemgrep: go.lang.security.audit.xss.import-text-template.import-text-template

	"github.com/amp-labs/connectors/common"
	"github.com/go-playground/validator"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	ErrProviderCatalogNotFound = errors.New("provider or provider catalog not found")
	ErrProviderOptionNotFound  = errors.New("provider option not found")
	ErrClient                  = errors.New("client creation failed")
)

func ReadCatalog() (CatalogType, error) {
	catalog, err := clone[CatalogType](catalog)
	if err != nil {
		return nil, err
	}

	// Validate the provider configuration
	v := validator.New()
	for provider, providerInfo := range catalog {
		if err := v.Struct(providerInfo); err != nil {
			return nil, err
		}

		providerInfo.Name = provider
	}

	return catalog, nil
}

// SetInfo sets the information for a specific provider in the catalog.
// This is useful to enable experimental providers or to override the default
// provider information. As a general rule, don't ever call this function
// in production code unless you have a compelling reason to do so.
func SetInfo(provider Provider, info ProviderInfo) {
	catalog[provider] = info
}

// ReadInfo reads the information from the catalog for specific provider. It also performs string substitution
// on the values in the config that are surrounded by {{}}.
func ReadInfo(provider Provider, substitutions *map[string]string) (*ProviderInfo, error) {
	pInfo, ok := catalog[provider]
	if !ok {
		return nil, ErrProviderCatalogNotFound
	}

	// Clone before modifying
	providerInfo, err := clone[ProviderInfo](pInfo)
	if err != nil {
		return nil, err
	}

	providerInfo.Name = provider

	// Validate the provider configuration
	v := validator.New()
	if err := v.Struct(providerInfo); err != nil {
		return nil, err
	}

	if substitutions == nil {
		substitutions = &map[string]string{}
	}

	// Apply substitutions to the provider configuration values which contain variables in the form of {{var}}.
	if err := substituteStruct(&providerInfo, substitutions); err != nil {
		return nil, err
	}

	return &providerInfo, nil
}

// substituteStruct performs string substitution on the fields of the input struct
// using the substitutions map.
func substituteStruct(input interface{}, substitutions *map[string]string) (err error) { //nolint:gocognit,cyclop,lll
	configStruct := reflect.ValueOf(input).Elem()
	for i := 0; i < configStruct.NumField(); i++ {
		field := configStruct.Field(i)

		// If the field is a string, perform substitution on it.
		if field.Kind() == reflect.String {
			substitutedVal, err := substitute(field.String(), substitutions)
			if err != nil {
				return err
			}

			field.SetString(substitutedVal)
		}

		if field.Kind() == reflect.Pointer {
			if field.Elem().Kind() == reflect.Struct {
				err := substituteStruct(field.Elem().Addr().Interface(), substitutions)
				if err != nil {
					return err
				}
			}
		}

		// If the field is a struct, perform substitution on its fields.
		if field.Kind() == reflect.Struct {
			err := substituteStruct(field.Addr().Interface(), substitutions)
			if err != nil {
				return err
			}
		}

		// If the field is a map, perform substitution on its values.
		if field.Kind() == reflect.Map {
			for _, key := range field.MapKeys() {
				val := field.MapIndex(key)
				if val.Kind() == reflect.String {
					substitutedVal, err := substitute(val.String(), substitutions)
					if err != nil {
						return err
					}

					field.SetMapIndex(key, reflect.ValueOf(substitutedVal))
				}
			}
		}
	}

	return nil
}

// substitute performs string substitution on the input string
// using the substitutions map.
func substitute(input string, substitutions *map[string]string) (string, error) {
	tmpl, err := template.New("-").Parse(input)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	err = tmpl.Execute(&result, substitutions)
	if err != nil {
		return "", err
	}

	return result.String(), nil
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
	Config *oauth2.Config
	Token  *oauth2.Token
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
	OAuth2ClientCreds *clientcredentials.Config

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
		if i.OauthOpts == nil {
			return nil, fmt.Errorf("%w: %s", ErrClient, "oauth2 options not found")
		}

		switch i.OauthOpts.GrantType {
		case AuthorizationCode:
			return createOAuth2AuthCodeHTTPClient(ctx, params.Client, params.Debug, params.OAuth2AuthCodeCreds)
		case ClientCredentials:
			return createOAuth2ClientCredentialsHTTPClient(ctx, params.Client, params.Debug, params.OAuth2ClientCreds)
		case PKCE:
			return nil, fmt.Errorf("%w: %s", ErrClient, "PKCE grant type not supported")
		default:
			return nil, fmt.Errorf("%w: unsupported grant type %q", ErrClient, i.OauthOpts.GrantType)
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
	cfg *clientcredentials.Config,
) (common.AuthenticatedHTTPClient, error) {
	options := []common.OAuthOption{
		common.WithOAuthClient(getClient(client)),
		common.WithTokenSource(cfg.TokenSource(ctx)),
	}

	if dbg {
		options = append(options, common.WithOAuthDebug(common.PrintRequestAndResponse))
	}

	oauthClient, err := common.NewOAuthHTTPClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create oauth2 client: %w", ErrClient, err)
	}

	return oauthClient, nil
}

func createApiKeyHTTPClient( //nolint:ireturn
	ctx context.Context,
	client *http.Client,
	dbg bool,
	info *ProviderInfo,
	apiKey string,
) (common.AuthenticatedHTTPClient, error) {
	if info.ApiKeyOpts.ValuePrefix != "" {
		apiKey = info.ApiKeyOpts.ValuePrefix + apiKey
	}

	opts := []common.HeaderAuthClientOption{
		common.WithHeaderClient(getClient(client)),
	}

	if dbg {
		opts = append(opts, common.WithHeaderDebug(common.PrintRequestAndResponse))
	}

	c, err := common.NewApiKeyAuthHTTPClient(ctx, info.ApiKeyOpts.HeaderName, apiKey, opts...)
	if err != nil {
		panic(err)
	}

	return c, nil
}

func pointerFor[T any](value T) *T {
	return &value
}

// clone uses gob to deep copy objects.
func clone[T any](input T) (T, error) { // nolint:ireturn
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(input); err != nil {
		return input, err
	}

	var clone T
	if err := dec.Decode(&clone); err != nil {
		return input, err
	}

	return clone, nil
}
