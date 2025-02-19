package providers

const (
	GoogleMeet Provider = "googleMeet"
)

//nolint:funlen
func init() {
	SetInfo(GoogleMeet, ProviderInfo{
		DisplayName: "Google Meet",
		AuthType:    Oauth2,
		BaseURL:     "https://meet.googleapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline", "prompt": "consent"},
			TokenURL:                  "https://oauth2.googleapis.com/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739948447/google-meet-icon-2020-_gngzfu.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739948447/google-meet-icon-2020-_gngzfu.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739948447/google-meet-icon-2020-_gngzfu.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739948447/google-meet-icon-2020-_gngzfu.svg",
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
