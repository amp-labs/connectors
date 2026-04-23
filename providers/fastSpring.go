package providers

const FastSpring Provider = "fastSpring"

func init() {
	// FastSpring configuration
	// API: https://developer.fastspring.com/reference/getting-started-with-your-api
	// Authentication: Basic Auth (Base64-encoded username:password from API Credentials in FastSpring app)
	SetInfo(FastSpring, ProviderInfo{
		DisplayName: "FastSpring",
		AuthType:    Basic,
		BaseURL:     "https://api.fastspring.com",
		BasicOpts: &BasicAuthOpts{
			DocsURL: "https://developer.fastspring.com/reference/getting-started-with-your-api",
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Delete:    false,
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773317338/media/fastspring.com_1773316963.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773317338/media/fastspring.com_1773316969.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773317338/media/fastspring.com_1773316963.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773317338/media/fastspring.com_1773317335.png",
			},
		},
	})
}
