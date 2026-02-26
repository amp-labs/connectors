package providers

import "github.com/amp-labs/connectors/common"

const Atlassian Provider = "atlassian"

const (
	// ModuleAtlassianJira is the module used for listing Jira issues.
	ModuleAtlassianJira common.ModuleID = "jira"
	// ModuleAtlassianConfluence is the module used for Atlassian Confluence.
	ModuleAtlassianConfluence common.ModuleID = "confluence"
)

// nolint:funlen
func init() {
	// Atlassian Configuration
	SetInfo(Atlassian, ProviderInfo{
		DisplayName: "Atlassian",
		AuthType:    Oauth2,
		BaseURL:     "https://api.atlassian.com/ex/confluence/50407e58-501b-44ed-8a86-c5e1e2fcf009/wiki/api",
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
			ModuleAtlassianConfluence: {
				BaseURL:     "https://api.atlassian.com/ex/confluence/{{.cloudId}}/wiki/api",
				DisplayName: "Atlassian Confluence",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
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
					// Jira and Confluence modules require it for their BaseURLs.
					// This is acquired using workspace.
					Name: "cloudId",
				},
			},
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "App name",
					DocsURL:     "https://support.atlassian.com/organization-administration/docs/update-your-product-and-site-url/",
					ModuleDependencies: &ModuleDependencies{
						ModuleAtlassianJira:       ModuleDependency{},
						ModuleAtlassianConfluence: ModuleDependency{},
					},
				},
			},
		},
	})
}
