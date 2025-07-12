package providers

// Avoma is the identifier for the Avoma provider.
const Avoma Provider = "avoma"

//nolint:lll
func init() {
	// Avoma configuration
	SetInfo(Avoma, ProviderInfo{
		DisplayName: "Avoma",
		AuthType:    ApiKey,
		BaseURL:     "https://api.avoma.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://help.avoma.com/api-integration-for-avoma",
		}, Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747987575/media/avoma.com_1747987574.jpg",
				LogoURL: "https://startupstorymedia.com/wp-content/uploads/2021/12/avoma.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747987575/media/avoma.com_1747987574.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1747987549/media/avoma.com_1747987547.png",
			},
		},
	})
}
