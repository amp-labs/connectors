package providers

const Dovetail Provider = "dovetail"

func init() {
	SetInfo(Dovetail, ProviderInfo{
		DisplayName: "Dovetail",
		AuthType:    ApiKey,
		BaseURL:     "https://dovetail.com/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://developers.dovetail.com/docs/authorization",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726578226/media/dovetail.com_1726578227.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726551927/media/dovetail.com_1726551926.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726551991/media/dovetail.com_1726551991.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726551927/media/dovetail.com_1726551926.svg",
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
