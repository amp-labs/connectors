package providers

const Freshdesk Provider = "freshdesk"

func init() {
	SetInfo(Freshdesk, ProviderInfo{
		DisplayName: "Freshdesk",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.freshdesk.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722321939/media/freshdesk_1722321938.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722321905/media/freshdesk_1722321903.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722321939/media/freshdesk_1722321938.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722321995/media/freshdesk_1722321994.svg",
			},
		},
	})
}
