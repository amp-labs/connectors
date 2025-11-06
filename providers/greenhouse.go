package providers

const Greenhouse Provider = "greenhouse"

//nolint:lll
func init() {
	// Greenhouse Configuration
	SetInfo(Greenhouse, ProviderInfo{
		DisplayName: "Greenhouse",
		AuthType:    Oauth2,
		BaseURL:     "https://harvest.greenhouse.io",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.greenhouse.io/authorize",
			AuthURLParams:             map[string]string{"state": "csrf_prevention_token_abc987"},
			TokenURL:                  "https://auth.greenhouse.io/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Delete: false,
				Insert: false,
				Update: false,
				Upsert: false,
			},
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760017955/media/developers.greenhouse.io_1760017960.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760020366/media/developers.greenhouse.io_1760020371.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760017955/media/developers.greenhouse.io_1760017960.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760018070/media/developers.greenhouse.io_1760018076.svg",
			},
		},
	})
}
