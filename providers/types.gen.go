// Package providers provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
package providers

// Defines values for ApiKeyOptsAttachmentType.
const (
	Header ApiKeyOptsAttachmentType = "header"
	Query  ApiKeyOptsAttachmentType = "query"
)

// Defines values for AuthType.
const (
	ApiKey AuthType = "apiKey"
	Basic  AuthType = "basic"
	None   AuthType = "none"
	Oauth2 AuthType = "oauth2"
)

// Defines values for Oauth2OptsGrantType.
const (
	AuthorizationCode Oauth2OptsGrantType = "authorizationCode"
	ClientCredentials Oauth2OptsGrantType = "clientCredentials"
	PKCE              Oauth2OptsGrantType = "PKCE"
	Password          Oauth2OptsGrantType = "password"
)

// ApiKeyOpts Configuration for API key. Must be provided if authType is apiKey.
type ApiKeyOpts struct {
	// AttachmentType How the API key should be attached to requests.
	AttachmentType ApiKeyOptsAttachmentType `json:"attachmentType" validate:"required"`

	// DocsURL URL with more information about how to get or use an API key.
	DocsURL string            `json:"docsURL,omitempty"`
	Header  *ApiKeyOptsHeader `json:"header,omitempty"`
	Query   *ApiKeyOptsQuery  `json:"query,omitempty"`
}

// ApiKeyOptsAttachmentType How the API key should be attached to requests.
type ApiKeyOptsAttachmentType string

// ApiKeyOptsHeader defines model for ApiKeyOptsHeader.
type ApiKeyOptsHeader struct {
	// Name The name of the header to be used for the API key.
	Name string `json:"name"`

	// ValuePrefix The prefix to be added to the API key value when it is sent in the header.
	ValuePrefix string `json:"valuePrefix,omitempty"`
}

// ApiKeyOptsQuery defines model for ApiKeyOptsQuery.
type ApiKeyOptsQuery struct {
	// Name The name of the query parameter to be used for the API key.
	Name string `json:"name"`
}

// AuthType defines model for AuthType.
type AuthType string

// BulkWriteSupport defines model for BulkWriteSupport.
type BulkWriteSupport struct {
	Delete bool `json:"delete"`
	Insert bool `json:"insert"`
	Update bool `json:"update"`
	Upsert bool `json:"upsert"`
}

// CatalogType defines model for CatalogType.
type CatalogType map[string]ProviderInfo

// Oauth2Opts Configuration for OAuth2.0. Must be provided if authType is oauth2.
type Oauth2Opts struct {
	// Audience A list of URLs that represent the audience for the token, which is needed for some client credential grant flows.
	Audience []string `json:"audience,omitempty"`

	// AuthURL The authorization URL.
	AuthURL       string            `json:"authURL,omitempty"`
	AuthURLParams map[string]string `json:"authURLParams,omitempty"`

	// DocsURL URL with more information about where to retrieve Client ID and Client Secret, etc.
	DocsURL string `json:"docsURL,omitempty"`

	// ExplicitScopesRequired Whether scopes are required to be known ahead of the OAuth flow.
	ExplicitScopesRequired bool `json:"explicitScopesRequired"`

	// ExplicitWorkspaceRequired Whether the workspace is required to be known ahead of the OAuth flow.
	ExplicitWorkspaceRequired bool                `json:"explicitWorkspaceRequired"`
	GrantType                 Oauth2OptsGrantType `json:"grantType"`
	TokenMetadataFields       TokenMetadataFields `json:"tokenMetadataFields"`

	// TokenURL The token URL.
	TokenURL string `json:"tokenURL" validate:"required"`
}

// Oauth2OptsGrantType defines model for Oauth2Opts.GrantType.
type Oauth2OptsGrantType string

// Provider defines model for Provider.
type Provider = string

// ProviderInfo defines model for ProviderInfo.
type ProviderInfo struct {
	// ApiKeyOpts Configuration for API key. Must be provided if authType is apiKey.
	ApiKeyOpts *ApiKeyOpts `json:"apiKeyOpts,omitempty"`
	AuthType   AuthType    `json:"authType" validate:"required"`

	// BaseURL The base URL for making API requests.
	BaseURL string `json:"baseURL" validate:"required"`

	// DisplayName The display name of the provider, if omitted, defaults to provider name.
	DisplayName string `json:"displayName,omitempty"`
	Name        string `json:"name"`

	// Oauth2Opts Configuration for OAuth2.0. Must be provided if authType is oauth2.
	Oauth2Opts *Oauth2Opts `json:"oauth2Opts,omitempty"`

	// PostAuthInfoNeeded If true, we require additional information after auth to start making requests.
	PostAuthInfoNeeded bool         `json:"postAuthInfoNeeded,omitempty"`
	ProviderOpts       ProviderOpts `json:"providerOpts"`
	Support            Support      `json:"support" validate:"required"`
}

// ProviderOpts defines model for ProviderOpts.
type ProviderOpts map[string]string

// Support defines model for Support.
type Support struct {
	BulkWrite BulkWriteSupport `json:"bulkWrite" validate:"required"`
	Proxy     bool             `json:"proxy"`
	Read      bool             `json:"read"`
	Subscribe bool             `json:"subscribe"`
	Write     bool             `json:"write"`
}

// TokenMetadataFields defines model for TokenMetadataFields.
type TokenMetadataFields struct {
	ConsumerRefField  string `json:"consumerRefField,omitempty"`
	ScopesField       string `json:"scopesField,omitempty"`
	WorkspaceRefField string `json:"workspaceRefField,omitempty"`
}
