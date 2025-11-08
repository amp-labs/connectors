package providers

import "github.com/amp-labs/connectors/common"

const Atlassian Provider = "atlassian"

const (
	// ModuleAtlassianJira is the module used for listing Jira issues.
	ModuleAtlassianJira common.ModuleID = "jira"
)

// nolint:funlen
func init() {
	// Atlassian Configuration
	SetInfo(Atlassian, ProviderInfo{
		DisplayName: "Atlassian",
		AuthType:    Oauth2,
		BaseURL:     "https://api.atlassian.com/ex/jira/{{.cloudId}}/rest/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true, // Needed for GetPostAuthInfo call
		},
		PostAuthInfoNeeded: true,
		DefaultModule:      ModuleAtlassianJira,
		Modules: &Modules{
			ModuleAtlassianJira: {
				BaseURL:     "https://api.atlassian.com/ex/jira/{{.cloudId}}/rest/api",
				DisplayName: "Atlassian Jira",
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
					ModuleDependencies: &ModuleDependencies{
						ModuleAtlassianJira: ModuleDependency{},
					},
				},
			},
		},
	})
}
