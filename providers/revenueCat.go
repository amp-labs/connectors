package providers

const RevenueCat Provider = "revenueCat"

func init() {
	// RevenueCat configuration
	SetInfo(RevenueCat, ProviderInfo{
		DisplayName: "RevenueCat",
		AuthType:    ApiKey,
		BaseURL:     "https://api.revenuecat.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://www.revenuecat.com/docs",
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1767679120/media/revenueCat_1767679120.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1767683245/media/revenueCat_1767683244.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1767679120/media/revenueCat_1767679120.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1767680783/media/revenueCat_1767680782.svg",
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
