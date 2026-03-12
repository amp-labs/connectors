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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773316914/media/fastspring.com_1773316911.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773316972/media/fastspring.com_1773316969.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773316914/media/fastspring.com_1773316911.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773316972/media/fastspring.com_1773316969.png",
			},
		},
	})
}
