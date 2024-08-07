package providers

const Maxio Provider = "maxio"

func init() {
	SetInfo(Maxio, ProviderInfo{
		DisplayName: "Maxio",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.chargify.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328568/media/maxio_1722328567.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328550/media/maxio_1722328549.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328600/media/maxio_1722328599.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328600/media/maxio_1722328599.svg",
			},
		},
	})
}
