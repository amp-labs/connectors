package providers

const Klaviyo Provider = "klaviyo"

func init() {
	// Klaviyo configuration
	SetInfo(Klaviyo, ProviderInfo{
		DisplayName: "Klaviyo",
		AuthType:    Oauth2,
		BaseURL:     "https://a.klaviyo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://www.klaviyo.com/oauth/authorize",
			TokenURL:                  "https://a.klaviyo.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480320/media/klaviyo_1722480318.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480320/media/klaviyo_1722480318.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480368/media/klaviyo_1722480367.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722480352/media/klaviyo_1722480351.svg",
			},
		},
	})
}
