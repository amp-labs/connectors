package providers

const Brex Provider = "brex"

func init() {
	// Brex configuration
	SetInfo(Brex, ProviderInfo{
		DisplayName: "Brex",
		AuthType:    ApiKey,
		BaseURL:     "https://platform.brexapis.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765886625/media/brex.com_1765886623.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765886646/media/brex.com_1765886646.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765886687/media/brex.com_1765886687.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765886662/media/brex.com_1765886662.svg",
			},
		},
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://developer.brex.com/docs/authentication/",
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
