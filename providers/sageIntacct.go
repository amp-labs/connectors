package providers

const SageIntacct Provider = "sageIntacct"

func init() {
	SetInfo(SageIntacct, ProviderInfo{
		DisplayName: "Sage Intacct",
		AuthType:    Oauth2,
		BaseURL:     "https://api.intacct.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.intacct.com/ia/api/v1/oauth2/authorize",
			TokenURL:                  "https://api.intacct.com/ia/api/v1/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			DocsURL:                   "https://developer.sage.com/intacct/docs/developer-portal/getting-started/quick-start",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508463/media/sage.com_1755508463.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508463/media/sage.com_1755508463.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508405/media/sage.com_1755508404.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1755508405/media/sage.com_1755508404.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
