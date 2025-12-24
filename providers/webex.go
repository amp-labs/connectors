package providers

const Webex Provider = "webex"

func init() {
	SetInfo(Webex, ProviderInfo{
		DisplayName: "Webex",
		AuthType:    Oauth2,
		BaseURL:     "https://webexapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://webexapis.com/v1/authorize",
			TokenURL:                  "https://webexapis.com/v1/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
			DocsURL: "https://developer.webex.com/docs/run-an-oauth-integration",
		},

		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765316683/media/webex.com_1765316683.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765316535/media/webex.com_1765316534.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765316683/media/webex.com_1765316683.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765316535/media/webex.com_1765316534.svg",
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
