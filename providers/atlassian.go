package providers

const Atlassian Provider = "atlassian"

func init() {
	// Atlassian Configuration
	SetInfo(Atlassian, ProviderInfo{
		DisplayName: "Atlassian (Jira, Confluence)",
		AuthType:    Oauth2,
		BaseURL:     "https://api.atlassian.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
