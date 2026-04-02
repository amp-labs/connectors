package providers

const HousecallPro Provider = "housecallPro"

func init() {
	SetInfo(HousecallPro, ProviderInfo{
		DisplayName: "Housecall Pro",
		AuthType:    ApiKey,
		BaseURL:     "https://api.housecallpro.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Token ",
			},
			DocsURL: "https://docs.housecallpro.com/docs/housecall-public-api/b87d37ae48a0d-authentication",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774023826/media/housecallpro.com_1774023825.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774023927/media/housecallpro.com_1774023927.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774023826/media/housecallpro.com_1774023825.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774023927/media/housecallpro.com_1774023927.svg",
			},
		},
	})
}
