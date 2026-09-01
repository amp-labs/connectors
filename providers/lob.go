package providers

const Lob = "lob"

func init() {
	SetInfo(Lob, ProviderInfo{
		DisplayName: "Lob",
		AuthType:    Basic,
		BaseURL:     "https://api.lob.com",
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787090538/lob_icon_smybnm.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787089526/media/lob_1787089525.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787090538/lob_icon_smybnm.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1787089543/media/lob_1787089543.svg",
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
			Write:     true,
		},
	})
}
