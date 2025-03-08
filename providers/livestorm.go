package providers

const (
	Livestorm Provider = "livestorm"
)

func init() {
	SetInfo(Livestorm, ProviderInfo{
		DisplayName: "Livestorm",
		AuthType:    Oauth2,
		BaseURL:     "https://api.livestorm.co",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.livestorm.co/oauth/authorize",
			TokenURL:                  "https://app.livestorm.co/oauth/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158143/media/api.livestorm.co_1741158142.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158176/media/api.livestorm.co_1741158174.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158111/media/api.livestorm.co_1741158108.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158160/media/api.livestorm.co_1741158158.svg",
			},
		},
	})
}
