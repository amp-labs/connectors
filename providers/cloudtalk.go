package providers

const CloudTalk Provider = "cloudTalk"

func init() {
	// CloudTalk Configuration
	SetInfo(CloudTalk, ProviderInfo{
		DisplayName: "CloudTalk",
		AuthType:    Basic,
		BaseURL:     "https://my.cloudtalk.io/api",
		BasicOpts: &BasicAuthOpts{
			DocsURL: "https://developers.cloudtalk.io/reference/authentication",
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     false,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765381242/media/cloudtalk.io_1765381242.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765381169/media/cloudtalk.io_1765381168.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765381242/media/cloudtalk.io_1765381242.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765381169/media/cloudtalk.io_1765381168.svg",
			},
		},
	})
}
