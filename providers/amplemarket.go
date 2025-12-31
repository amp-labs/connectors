package providers

const Amplemarket Provider = "amplemarket"

func init() {
	SetInfo(Amplemarket, ProviderInfo{
		DisplayName: "Amplemarket",
		AuthType:    ApiKey,
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.amplemarket.com/guides/quickstart",
		},
		BaseURL: "https://api.amplemarket.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766067606/media/amplemarket.com_1766067602.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766067649/media/amplemarket.com_1766067648.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766067629/media/amplemarket.com_1766067627.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766067671/media/amplemarket.com_1766067668.svg",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
