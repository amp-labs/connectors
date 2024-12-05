package providers

const GoToWebinar Provider = "goToWebinar"

func init() {
	SetInfo(GoToWebinar, ProviderInfo{
		DisplayName: "GoToWebinar",
		AuthType:    Oauth2,
		BaseURL:     "https://api.getgo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://authentication.logmeininc.com/oauth/authorize",
			TokenURL:                  "https://authentication.logmeininc.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581742/media/goto.com_1731581740.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581742/media/goto.com_1731581740.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581774/media/goto.com_1731581772.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1731581774/media/goto.com_1731581772.svg",
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
