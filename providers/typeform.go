package providers

const Typeform Provider = "typeform"

func init() {
	SetInfo(Typeform, ProviderInfo{
		DisplayName: "Typeform",
		AuthType:    Oauth2,
		BaseURL:     "https://api.typeform.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.typeform.com/oauth/authorize",
			TokenURL:                  "https://api.typeform.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722183939/media/typeform_1722183939.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184066/media/typeform_1722184065.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722183939/media/typeform_1722183939.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722184066/media/typeform_1722184065.jpg",
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
