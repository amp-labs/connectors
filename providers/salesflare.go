package providers

const Salesflare Provider = "salesflare"

func init() {
	// Salesflare configuration
	SetInfo(Salesflare, ProviderInfo{
		DisplayName: "Salesflare",
		AuthType:    ApiKey,
		BaseURL:     "https://api.salesflare.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://api.salesflare.com/docs#section/Introduction/Authentication",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722457532/media/salesflare_1722457532.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722457532/media/salesflare_1722457532.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722457496/media/salesflare_1722457495.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722457496/media/salesflare_1722457495.png",
			},
		},
	})
}
