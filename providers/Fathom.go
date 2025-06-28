package providers

const Fathom Provider = "fathom"

func init() {
	SetInfo(Fathom, ProviderInfo{
		DisplayName: "Fathom",
		AuthType:    ApiKey,
		BaseURL:     "https://api.fathom.ai",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Api-Key",
			},
			DocsURL: "https://docs.fathom.ai",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749723319/media/fathom.video_1749723318.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749723372/media/fathom.video_1749723371.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749723319/media/fathom.video_1749723318.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749723350/media/fathom.video_1749723350.svg",
			},
		},
	})
}
