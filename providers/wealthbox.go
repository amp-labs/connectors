package providers

const Wealthbox Provider = "wealthbox"

func init() {
	SetInfo(Wealthbox, ProviderInfo{
		DisplayName: "Wealthbox",
		AuthType:    ApiKey,
		BaseURL:     "https://api.crmworkspace.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "ACCESS_TOKEN",
			},
			DocsURL: "https://dev.wealthbox.com/#topics-authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782311753/media/wealthbox.com_1782311752.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782311574/media/wealthbox.com_1782311569.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782311753/media/wealthbox.com_1782311752.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782311574/media/wealthbox.com_1782311569.svg",
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
