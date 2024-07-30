package providers

const Freshservice Provider = "freshservice"

func init() {
	SetInfo(Freshservice, ProviderInfo{
		DisplayName: "Freshservice",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.freshservice.com",
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
		PostAuthInfoNeeded: false,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326054/media/freshservice_1722326053.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326028/media/freshservice_1722326026.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326054/media/freshservice_1722326053.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326085/media/freshservice_1722326084.svg",
			},
		},
	})
}
