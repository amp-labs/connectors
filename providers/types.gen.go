// Package providers provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package providers

// Defines values for AuthType.
const (
	Oauth2 AuthType = "oauth2"
)

// AuthType defines model for AuthType.
type AuthType string

// CatalogType defines model for CatalogType.
type CatalogType map[string]ProviderInfo

// OauthOpts defines model for OauthOpts.
type OauthOpts struct {
	AuthURL                   string              `json:"authURL" validate:"required"`
	ExplicitScopesRequired    bool                `json:"explicitScopesRequired"`
	ExplicitWorkspaceRequired bool                `json:"explicitWorkspaceRequired"`
	TokenMetadataFields       TokenMetadataFields `json:"tokenMetadataFields"`
	TokenURL                  string              `json:"tokenURL" validate:"required"`
}

// Provider defines model for Provider.
type Provider = string

// ProviderInfo defines model for ProviderInfo.
type ProviderInfo struct {
	AuthType     AuthType     `json:"authType" validate:"required"`
	BaseURL      string       `json:"baseURL" validate:"required"`
	OauthOpts    OauthOpts    `json:"oauthOpts" validate:"required"`
	ProviderOpts ProviderOpts `json:"providerOpts"`
	Support      Support      `json:"support" validate:"required"`
}

// ProviderOpts defines model for ProviderOpts.
type ProviderOpts map[string]string

// Support defines model for Support.
type Support struct {
	BulkWrite bool `json:"bulkWrite"`
	Proxy     bool `json:"proxy"`
	Read      bool `json:"read"`
	Subscribe bool `json:"subscribe"`
	Write     bool `json:"write"`
}

// TokenMetadataFields defines model for TokenMetadataFields.
type TokenMetadataFields struct {
	ConsumerRefField  string `json:"consumerRefField,omitempty"`
	ScopesField       string `json:"scopesField,omitempty"`
	WorkspaceRefField string `json:"workspaceRefField,omitempty"`
}
