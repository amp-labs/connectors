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
			DocsURL: "https://www.revenuecat.com/docs/api-v2",
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://www.revenuecat.com/docs/img/logo-rc-small.svg",
				LogoURL: "https://www.revenuecat.com/docs/img/logo-rc-small.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://www.revenuecat.com/docs/img/logo-rc-small.svg",
				LogoURL: "https://www.revenuecat.com/docs/img/logo-rc-small.svg",
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
