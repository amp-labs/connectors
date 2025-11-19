// nolint
package substitutions

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/stretchr/testify/assert"
)

// TestSubstitute tests the basic string substitution functionality.
func TestSubstitute(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		substitutions map[string]string
		expected      string
		expectError   bool
	}{
		{
			name:          "simple substitution",
			input:         "https://{{.workspace}}.my.salesforce.com",
			substitutions: map[string]string{"workspace": "test"},
			expected:      "https://test.my.salesforce.com",
			expectError:   false,
		},
		{
			name:          "multiple substitutions",
			input:         "https://{{.workspace}}.{{.region}}.amazonaws.com",
			substitutions: map[string]string{"workspace": "test", "region": "us-west-2"},
			expected:      "https://test.us-west-2.amazonaws.com",
			expectError:   false,
		},
		{
			name:          "missing variable",
			input:         "https://{{.workspace}}.my.salesforce.com",
			substitutions: map[string]string{},
			expected:      "",
			expectError:   true,
		},
		{
			name:          "no substitutions needed",
			input:         "https://api.atlassian.com",
			substitutions: map[string]string{"workspace": "test"},
			expected:      "https://api.atlassian.com",
			expectError:   false,
		},
		{
			name:          "invalid template",
			input:         "https://{{.workspace.my.salesforce.com",
			substitutions: map[string]string{"workspace": "test"},
			expected:      "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := substitute(tt.input, tt.substitutions)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestSubstituteStruct tests the struct substitution functionality.
func TestSubstituteStruct(t *testing.T) {
	type NestedStruct struct {
		URL string
	}

	type TestStruct struct {
		BaseURL   string
		AuthURL   string
		TokenURL  string
		Nested    NestedStruct
		NestedPtr *NestedStruct
		URLMap    map[string]string
		NonString int
	}

	tests := []struct {
		name          string
		input         *TestStruct
		substitutions map[string]string
		expected      *TestStruct
		expectError   bool
	}{
		{
			name: "complete substitution",
			input: &TestStruct{
				BaseURL:  "https://{{.workspace}}.my.salesforce.com",
				AuthURL:  "https://{{.workspace}}.my.salesforce.com/services/oauth2/authorize",
				TokenURL: "https://{{.workspace}}.my.salesforce.com/services/oauth2/token",
				Nested: NestedStruct{
					URL: "https://{{.workspace}}.my.salesforce.com/api",
				},
				NestedPtr: &NestedStruct{
					URL: "https://{{.workspace}}.my.salesforce.com/api/v2",
				},
				URLMap: map[string]string{
					"test": "https://{{.workspace}}.my.salesforce.com/test",
				},
				NonString: 42,
			},
			substitutions: map[string]string{
				"workspace": "test",
			},
			expected: &TestStruct{
				BaseURL:  "https://test.my.salesforce.com",
				AuthURL:  "https://test.my.salesforce.com/services/oauth2/authorize",
				TokenURL: "https://test.my.salesforce.com/services/oauth2/token",
				Nested: NestedStruct{
					URL: "https://test.my.salesforce.com/api",
				},
				NestedPtr: &NestedStruct{
					URL: "https://test.my.salesforce.com/api/v2",
				},
				URLMap: map[string]string{
					"test": "https://test.my.salesforce.com/test",
				},
				NonString: 42,
			},
			expectError: false,
		},
		{
			name: "missing variable",
			input: &TestStruct{
				BaseURL: "https://{{.workspace}}.my.salesforce.com",
			},
			substitutions: map[string]string{},
			expected:      nil,
			expectError:   true,
		},
		{
			name: "no substitutions needed",
			input: &TestStruct{
				BaseURL: "https://api.atlassian.com",
			},
			substitutions: map[string]string{"workspace": "test"},
			expected: &TestStruct{
				BaseURL: "https://api.atlassian.com",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := substituteStruct(tt.input, tt.substitutions)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.input)
			}
		})
	}
}

// TestProviderInfoSubstitution tests substitution with a real ProviderInfo-like structure.
func TestProviderInfoSubstitution(t *testing.T) {
	tests := []struct {
		name          string
		input         *ProviderInfo
		substitutions map[string]string
		expected      *ProviderInfo
		expectError   bool
	}{
		{
			name:  "provider info substitution",
			input: &atlassianProviderInfoPreSubstitution,
			substitutions: map[string]string{
				"workspace": "test-workspace",
				"cloudId":   "test-cloud-id",
				"metadata":  "test-metadata",
			},
			expected:    &atlasssianProviderInfoPostSubstitution,
			expectError: false,
		},
		{
			name:  "salesforce provider info substitution",
			input: &salesforcePresubstitution,
			substitutions: map[string]string{
				"workspace": "test-workspace",
			},
			expected:    &salesforcePostSubstitution,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := substituteStruct(tt.input, tt.substitutions)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tt.input)
			}
		})
	}
}

////// =============================================
////// salesforce & atlassian providerInfo types
////// =============================================

var (
	salesforcePresubstitution = ProviderInfo{
		DisplayName: "Salesforce",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.my.salesforce.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://{{.workspace}}.my.salesforce.com/services/oauth2/userinfo",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.my.salesforce.com/services/oauth2/authorize",
			TokenURL:                  "https://{{.workspace}}.my.salesforce.com/services/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "id",
				WorkspaceRefField: "instance_url",
				ScopesField:       "scope",
			},
			ScopeMappings: map[string]string{
				"test_scope": "https://{{.workspace}}.my.salesforce.com/test_scope",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: true,
				Delete: true,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
					DocsURL:     "https://help.salesforce.com/s/articleView?language=en_US&id=sf.faq_domain_name_what.htm&type=5",
				},
			},
		},
	}

	atlassianProviderInfoPreSubstitution = ProviderInfo{
		DisplayName: "Atlassian",
		AuthType:    Oauth2,
		BaseURL:     "https://api.{{.metadata}}.atlassian.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true, // Needed for GetPostAuthInfo call
		},
		PostAuthInfoNeeded: true,
		DefaultModule:      ModuleAtlassianJira,
		Modules: Modules{
			ModuleAtlassianJira: {
				BaseURL:     "https://api.atlassian.com/ex/jira/{{.cloudId}}/rest/api/3",
				DisplayName: "Atlassian Jira",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleAtlassianJiraConnect: {
				BaseURL:     "https://{{.workspace}}.atlassian.net/rest/api/3",
				DisplayName: "Atlassian Connect",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490152/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490153.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490205/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490206.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490152/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490153.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490205/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490206.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "cloudId",
				},
			},
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "App name",
					DocsURL:     "https://support.atlassian.com/organization-administration/docs/update-your-product-and-site-url/",
				},
			},
		},
	}
)

var (
	salesforcePostSubstitution = ProviderInfo{
		DisplayName: "Salesforce",
		AuthType:    Oauth2,
		BaseURL:     "https://test-workspace.my.salesforce.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://test-workspace.my.salesforce.com/services/oauth2/userinfo",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://test-workspace.my.salesforce.com/services/oauth2/authorize",
			TokenURL:                  "https://test-workspace.my.salesforce.com/services/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "id",
				WorkspaceRefField: "instance_url",
				ScopesField:       "scope",
			},
			ScopeMappings: map[string]string{
				"test_scope": "https://test-workspace.my.salesforce.com/test_scope",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: true,
				Delete: true,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
					DocsURL:     "https://help.salesforce.com/s/articleView?language=en_US&id=sf.faq_domain_name_what.htm&type=5",
				},
			},
		},
	}

	atlasssianProviderInfoPostSubstitution = ProviderInfo{
		DisplayName: "Atlassian",
		AuthType:    Oauth2,
		BaseURL:     "https://api.test-metadata.atlassian.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true, // Needed for GetPostAuthInfo call
		},
		PostAuthInfoNeeded: true,
		DefaultModule:      ModuleAtlassianJira,
		Modules: Modules{
			ModuleAtlassianJira: {
				BaseURL:     "https://api.atlassian.com/ex/jira/test-cloud-id/rest/api/3",
				DisplayName: "Atlassian Jira",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleAtlassianJiraConnect: {
				BaseURL:     "https://test-workspace.atlassian.net/rest/api/3",
				DisplayName: "Atlassian Connect",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490152/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490153.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490205/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490206.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490152/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490153.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722490205/media/const%20Atlassian%20Provider%20%3D%20%22atlassian%22_1722490206.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "cloudId",
				},
			},
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "App name",
					DocsURL:     "https://support.atlassian.com/organization-administration/docs/update-your-product-and-site-url/",
				},
			},
		},
	}
)

////// =============================================
////// copied from providers/types.gen.go
////// =============================================

// Defines values for ApiKeyAsBasicOptsFieldUsed.
const (
	PasswordField ApiKeyAsBasicOptsFieldUsed = "password"
	UsernameField ApiKeyAsBasicOptsFieldUsed = "username"

	ModuleAtlassianJira        common.ModuleID = "atlassianJira"
	ModuleAtlassianJiraConnect common.ModuleID = "atlassianJiraConnect"
)

// Defines values for ApiKeyOptsAttachmentType.
const (
	Header ApiKeyOptsAttachmentType = "header"
	Query  ApiKeyOptsAttachmentType = "query"
)

// Defines values for AuthType.
const (
	ApiKey AuthType = "apiKey"
	Basic  AuthType = "basic"
	Jwt    AuthType = "jwt"
	None   AuthType = "none"
	Oauth2 AuthType = "oauth2"
)

// Defines values for Oauth2OptsGrantType.
const (
	AuthorizationCode     Oauth2OptsGrantType = "authorizationCode"
	AuthorizationCodePKCE Oauth2OptsGrantType = "authorizationCodePKCE"
	ClientCredentials     Oauth2OptsGrantType = "clientCredentials"
	Password              Oauth2OptsGrantType = "password"
)

// Defines values for SubscribeOptsRegistrationTiming.
const (
	SubscribeOptsRegistrationTimingInstallation SubscribeOptsRegistrationTiming = "installation"
	SubscribeOptsRegistrationTimingIntegration  SubscribeOptsRegistrationTiming = "integration"
	SubscribeOptsRegistrationTimingProviderApp  SubscribeOptsRegistrationTiming = "providerApp"
)

// Defines values for SubscribeOptsSubscriptionScope.
const (
	SubscribeOptsSubscriptionScopeInstallation SubscribeOptsSubscriptionScope = "installation"
	SubscribeOptsSubscriptionScopeIntegration  SubscribeOptsSubscriptionScope = "integration"
)

// ApiKeyAsBasicOpts when this object is present, it means that this provider uses Basic Auth to actually collect an API key.
type ApiKeyAsBasicOpts struct {
	// FieldUsed whether the API key should be used as the username or password.
	FieldUsed ApiKeyAsBasicOptsFieldUsed `json:"fieldUsed,omitempty"`

	// KeyFormat How to transform the API key in to a basic auth user:pass string. The %s is replaced with the API key value.
	KeyFormat string `json:"keyFormat,omitempty"`
}

// ApiKeyAsBasicOptsFieldUsed whether the API key should be used as the username or password.
type ApiKeyAsBasicOptsFieldUsed string

// ApiKeyOpts Configuration for API key. Must be provided if authType is apiKey.
type ApiKeyOpts struct {
	// AttachmentType How the API key should be attached to requests.
	AttachmentType ApiKeyOptsAttachmentType `json:"attachmentType" validate:"required"`

	// DocsURL URL with more information about how to get or use an API key.
	DocsURL string `json:"docsURL,omitempty"`

	// Header Configuration for API key in header. Must be provided if type is in-header.
	Header *ApiKeyOptsHeader `json:"header,omitempty"`

	// Query Configuration for API key in query parameter. Must be provided if type is in-query.
	Query *ApiKeyOptsQuery `json:"query,omitempty"`
}

// ApiKeyOptsAttachmentType How the API key should be attached to requests.
type ApiKeyOptsAttachmentType string

// ApiKeyOptsHeader Configuration for API key in header. Must be provided if type is in-header.
type ApiKeyOptsHeader struct {
	// Name The name of the header to be used for the API key.
	Name string `json:"name"`

	// ValuePrefix The prefix to be added to the API key value when it is sent in the header.
	ValuePrefix string `json:"valuePrefix,omitempty"`
}

// ApiKeyOptsQuery Configuration for API key in query parameter. Must be provided if type is in-query.
type ApiKeyOptsQuery struct {
	// Name The name of the query parameter to be used for the API key.
	Name string `json:"name"`
}

// AuthHealthCheck A URL to check the health of a provider's credentials. It's used to see if the credentials are valid and if the provider is reachable.
type AuthHealthCheck struct {
	// Method The HTTP method to use for the health check. If not set, defaults to GET.
	Method string `json:"method,omitempty"`

	// SuccessStatusCodes The HTTP status codes that indicate a successful health check. If not set, defaults to 200 and 204.
	SuccessStatusCodes []int `json:"successStatusCodes,omitempty"`

	// Url a no-op URL to check the health of the credentials. The URL MUST not mutate any state. If the provider doesn't have such an endpoint, then don't provide credentialsHealthCheck.
	Url string `json:"url"`
}

// AuthType The type of authentication required by the provider.
type AuthType string

// BasicAuthOpts Configuration for Basic Auth. Optional.
type BasicAuthOpts struct {
	// ApiKeyAsBasic If true, the provider uses an API key which then gets encoded as a basic auth user:pass string.
	ApiKeyAsBasic bool `json:"apiKeyAsBasic,omitempty"`

	// ApiKeyAsBasicOpts when this object is present, it means that this provider uses Basic Auth to actually collect an API key
	ApiKeyAsBasicOpts *ApiKeyAsBasicOpts `json:"apiKeyAsBasicOpts,omitempty"`

	// DocsURL URL with more information about how to get or use an API key.
	DocsURL string `json:"docsURL,omitempty"`
}

// BulkWriteSupport defines model for BulkWriteSupport.
type BulkWriteSupport struct {
	Delete bool `json:"delete"`
	Insert bool `json:"insert"`
	Update bool `json:"update"`
	Upsert bool `json:"upsert"`
}

// CatalogType defines model for CatalogType.
type CatalogType map[string]ProviderInfo

// CatalogWrapper defines model for CatalogWrapper.
type CatalogWrapper struct {
	Catalog CatalogType `json:"catalog"`

	// Timestamp An RFC3339 formatted timestamp of when the catalog was generated.
	Timestamp string `json:"timestamp" validate:"required"`
}

// Labels defines model for Labels.
type Labels map[string]string

// Media defines model for Media.
type Media struct {
	// DarkMode Media to be used in dark mode.
	DarkMode *MediaTypeDarkMode `json:"darkMode,omitempty"`

	// Regular Media for light/regular mode.
	Regular *MediaTypeRegular `json:"regular,omitempty"`
}

// MediaTypeDarkMode Media to be used in dark mode.
type MediaTypeDarkMode struct {
	// IconURL URL to the icon for the provider that is to be used in dark mode.
	IconURL string `json:"iconURL,omitempty"`

	// LogoURL URL to the logo for the provider that is to be used in dark mode.
	LogoURL string `json:"logoURL,omitempty"`
}

// MediaTypeRegular Media for light/regular mode.
type MediaTypeRegular struct {
	// IconURL URL to the icon for the provider.
	IconURL string `json:"iconURL,omitempty"`

	// LogoURL URL to the logo for the provider.
	LogoURL string `json:"logoURL,omitempty"`
}

// MetadataItemInput defines model for MetadataItemInput.
type MetadataItemInput struct {
	// DisplayName The human-readable name for the field
	DisplayName string `json:"displayName,omitempty"`

	// DocsURL URL with more information about how to locate this value
	DocsURL string `json:"docsURL,omitempty"`

	// Name The internal identifier for the metadata field
	Name string `json:"name"`
}

// MetadataItemPostAuthentication defines model for MetadataItemPostAuthentication.
type MetadataItemPostAuthentication struct {
	// Name The internal identifier for the metadata field
	Name string `json:"name"`
}

// ModuleInfo defines model for ModuleInfo.
type ModuleInfo struct {
	BaseURL     string `json:"baseURL"`
	DisplayName string `json:"displayName"`

	// Support The supported features for the provider.
	Support Support `json:"support" validate:"required"`
}

// Modules The registry of provider modules.
type Modules = map[common.ModuleID]ModuleInfo

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

	// TokenMetadataFields Fields to be used to extract token metadata from the token response.
	TokenMetadataFields TokenMetadataFields `json:"tokenMetadataFields"`

	// ScopeMappings Maps input scopes to their full OAuth scope values with template variable support. Scopes not in this map are passed through unchanged. Needed for some providers.
	ScopeMappings map[string]string `json:"scopeMappings,omitempty"`

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

	// AuthHealthCheck A URL to check the health of a provider's credentials. It's used to see if the credentials are valid and if the provider is reachable.
	AuthHealthCheck *AuthHealthCheck `json:"authHealthCheck,omitempty"`

	// AuthType The type of authentication required by the provider.
	AuthType AuthType `json:"authType" validate:"required"`

	// BaseURL The base URL for making API requests.
	BaseURL string `json:"baseURL" validate:"required"`

	// BasicOpts Configuration for Basic Auth. Optional.
	BasicOpts     *BasicAuthOpts  `json:"basicOpts,omitempty"`
	DefaultModule common.ModuleID `json:"defaultModule"`

	// DisplayName The display name of the provider, if omitted, defaults to provider name.
	DisplayName string  `json:"displayName,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`
	Media       *Media  `json:"media,omitempty"`

	// Metadata Provider metadata that needs to be given by the user or fetched by the connector post authentication for the connector to work.
	Metadata *ProviderMetadata `json:"metadata,omitempty"`

	// Modules The registry of provider modules.
	Modules Modules `json:"modules,omitempty"`
	Name    string  `json:"name"`

	// Oauth2Opts Configuration for OAuth2.0. Must be provided if authType is oauth2.
	Oauth2Opts *Oauth2Opts `json:"oauth2Opts,omitempty"`

	// PostAuthInfoNeeded If true, we require additional information after auth to start making requests.
	PostAuthInfoNeeded bool `json:"postAuthInfoNeeded,omitempty"`

	// ProviderOpts Additional provider-specific metadata.
	ProviderOpts  ProviderOpts   `json:"providerOpts"`
	SubscribeOpts *SubscribeOpts `json:"subscribeOpts,omitempty"`

	// Support The supported features for the provider.
	Support Support `json:"support" validate:"required"`
}

// ProviderMetadata Provider metadata that needs to be given by the user or fetched by the connector post authentication for the connector to work.
type ProviderMetadata struct {
	// Input Metadata provided as manual input
	Input []MetadataItemInput `json:"input,omitempty"`

	// PostAuthentication Metadata fetched by the connector post authentication
	PostAuthentication []MetadataItemPostAuthentication `json:"postAuthentication,omitempty"`
}

// ProviderOpts Additional provider-specific metadata.
type ProviderOpts map[string]string

// SubscribeOpts defines model for SubscribeOpts.
type SubscribeOpts struct {
	// RegistrationTiming The timing of the registration.
	RegistrationTiming SubscribeOptsRegistrationTiming `json:"registrationTiming"`

	// SubscriptionScope The scope of the subscription.
	SubscriptionScope SubscribeOptsSubscriptionScope `json:"subscriptionScope"`

	// TargetURLScope The scope of the target URL.
	TargetURLScope interface{} `json:"targetURLScope"`
}

// SubscribeOptsRegistrationTiming The timing of the registration.
type SubscribeOptsRegistrationTiming string

// SubscribeOptsSubscriptionScope The scope of the subscription.
type SubscribeOptsSubscriptionScope string

// SubscribeSupport defines model for SubscribeSupport.
type SubscribeSupport struct {
	Create      *bool `json:"create,omitempty"`
	Delete      *bool `json:"delete,omitempty"`
	PassThrough *bool `json:"passThrough,omitempty"`
	Update      *bool `json:"update,omitempty"`
}

// Support The supported features for the provider.
type Support struct {
	BulkWrite        BulkWriteSupport  `json:"bulkWrite"                  validate:"required"`
	Proxy            bool              `json:"proxy"`
	Read             bool              `json:"read"`
	Subscribe        bool              `json:"subscribe"`
	SubscribeSupport *SubscribeSupport `json:"subscribeSupport,omitempty"`
	Write            bool              `json:"write"`
}

// TokenMetadataFields Fields to be used to extract token metadata from the token response.
type TokenMetadataFields struct {
	ConsumerRefField string `json:"consumerRefField,omitempty"`

	// OtherFields Additional fields to extract and transform from the token response
	OtherFields       *TokenMetadataFieldsOtherFields `json:"otherFields,omitempty"`
	ScopesField       string                          `json:"scopesField,omitempty"`
	WorkspaceRefField string                          `json:"workspaceRefField,omitempty"`
}

// TokenMetadataFieldsOtherFields Additional fields to extract and transform from the token response.
type TokenMetadataFieldsOtherFields = []struct {
	// Capture A regex expression to capture the value that we need from the path. There must be only one capture group named 'result' in the expression. If not provided, will cause an error.
	Capture string `json:"capture,omitempty"`

	// DisplayName The human-readable name of the field
	DisplayName string `json:"displayName"`

	// Name The internal name of the field
	Name string `json:"name"`

	// Path The path to the field in the token response (accepts dot notation for nested fields)
	Path string `json:"path"`
}
