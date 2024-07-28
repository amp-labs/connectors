package providers

const Chargebee Provider = "chargebee"

func init() {
	// Chargebee connfiguration
	// workspace maps to site
	SetInfo(Chargebee, ProviderInfo{
		DisplayName: "Chargebee",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.chargebee.com/api",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165406/media/chargebee.com_1722165405.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165406/media/chargebee.com_1722165405.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165406/media/chargebee.com_1722165405.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165406/media/chargebee.com_1722165405.jpg",
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
		PostAuthInfoNeeded: false,
	})
}
