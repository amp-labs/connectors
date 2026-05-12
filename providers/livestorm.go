package providers

const (
	Livestorm Provider = "livestorm"
)

func init() {
	// API token auth is used for development and integration tests until OAuth2 client
	// credentials are available from Livestorm support; then restore AuthType Oauth2 and
	// Oauth2Opts. https://developers.livestorm.co/docs/api-token-authentication
	SetInfo(Livestorm, ProviderInfo{
		DisplayName: "Livestorm",
		AuthType:    ApiKey,
		BaseURL:     "https://api.livestorm.co",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
				// Livestorm sends the API token as the full Authorization header value (no Bearer prefix).
			},
			DocsURL: "https://developers.livestorm.co/docs/api-token-authentication",
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
