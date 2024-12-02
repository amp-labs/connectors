package providers

const Lemlist Provider = "Lemlist"

func init() {
	SetInfo(Lemlist, ProviderInfo{
		DisplayName: "Lemlist",
		AuthType:    ApiKey,
		BaseURL:     "https://api.lemlist.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Query,
			Query: &ApiKeyOptsQuery{
				Name: "access_token",
			},
			DocsURL: "https://developer.lemlist.com",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1732702066/media/lemlist.com_1732702064.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1732702144/media/lemlist.com_1732702144.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1732702066/media/lemlist.com_1732702064.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1732702144/media/lemlist.com_1732702144.svg",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
