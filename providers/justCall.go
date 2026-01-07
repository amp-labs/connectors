package providers

const JustCall Provider = "justCall"

func init() {
	// JustCall Configuration
	SetInfo(JustCall, ProviderInfo{
		DisplayName: "JustCall",
		AuthType:    ApiKey,
		BaseURL:     "https://api.justcall.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
			},
			DocsURL: "https://developer.justcall.io/reference/authentication",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765374822/media/justcall.io_1765374817.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765374913/media/justcall.io_1765374911.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765374822/media/justcall.io_1765374817.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765374913/media/justcall.io_1765374911.jpg",
			},
		},
	})
}
