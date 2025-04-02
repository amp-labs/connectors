package providers

const Atlassian Provider = "atlassian"

const (
	// ModuleAtlassianJira is the module used for listing Jira issues.
	ModuleAtlassianJira string = "jira"
	// ModuleAtlassianJiraConnect is the module used for Atlassian Connect.
	ModuleAtlassianJiraConnect string = "atlassian-connect"
)

func init() {
	// Atlassian Configuration
	SetInfo(Atlassian, ProviderInfo{
		DisplayName: "Atlassian",
		AuthType:    Oauth2,
		BaseURL:     "https://api.atlassian.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true, // Needed for GetPostAuthInfo call
		},
		PostAuthInfoNeeded: true,
		Modules: &ModuleInfo{
			ModuleAtlassianJira: {
				BaseURL:     "https://api.atlassian.com/ex/jira/{{.cloudId}}/rest/api/3",
				DisplayName: "Jira",
				Support: ModuleSupport{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleAtlassianJiraConnect: {
				BaseURL:     "https://{{.workspace}}.atlassian.net/rest/api/3",
				DisplayName: "Atlassian Connect",
				Support: ModuleSupport{
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
	})
}
