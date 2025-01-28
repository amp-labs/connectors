package providers

const ChiliPiper Provider = "chilipiper"

func init() {
	SetInfo(ChiliPiper, ProviderInfo{
		DisplayName: "Chili Piper",
		AuthType:    ApiKey,
		BaseURL:     "https://fire.chilipiper.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737706607/media/chilipiper.com_1737706605.svg",
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
