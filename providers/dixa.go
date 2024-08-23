package providers

const Dixa Provider = "dixa"

func init() {
	SetInfo(Dixa, ProviderInfo{
		DisplayName: "Dixa",
		AuthType:    ApiKey,
		BaseURL:     "https://dev.dixa.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
			},
			DocsURL: "https://docs.dixa.io/docs/api-standards-rules/#authentication",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327729/media/dixa_1722327728.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724368834/media/p8slnkqpz9crzhrxenvj.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724367155/media/wrb7tnh66eaq0746rmqe.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327746/media/dixa_1722327745.svg",
			},
		},
	})
}
