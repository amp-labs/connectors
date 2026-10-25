package providers

const Orb Provider = "orb"

func init() {
	SetInfo(Orb, ProviderInfo{
		DisplayName: "Orb",
		AuthType:    ApiKey,
		// API paths are versioned under /v1, e.g. https://api.withorb.com/v1/customers.
		BaseURL: "https://api.withorb.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.withorb.com/essentials/authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792887357/media/orb_dark_icon.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792887359/media/orb_dark_logo.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792887362/media/orb_regular_icon.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1792887363/media/orb_regular_logo.svg",
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
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
