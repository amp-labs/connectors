package providers

const Formstack Provider = "formstack"

func init() {
	// Formstack configuration
	SetInfo(Formstack, ProviderInfo{
		DisplayName: "Formstack",
		AuthType:    Oauth2,
		BaseURL:     "https://www.formstack.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.formstack.com/api/v2/oauth2/authorize",
			TokenURL:                  "https://www.formstack.com/api/v2/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user_id",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062850/media/formstack_1722062849.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062824/media/formstack_1722062823.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062850/media/formstack_1722062849.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722062824/media/formstack_1722062823.svg",
			},
		},
	})
}
