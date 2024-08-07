package providers

const (
	Google Provider = "google"
)

//nolint:funlen
func init() {
	SetInfo(Google, ProviderInfo{
		DisplayName: "Google",
		AuthType:    Oauth2,
		BaseURL:     "https://www.googleapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://oauth2.googleapis.com/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349084/media/google_1722349084.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349053/media/google_1722349052.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349084/media/google_1722349084.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349053/media/google_1722349052.svg",
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
