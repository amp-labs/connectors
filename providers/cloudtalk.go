package providers

const CloudTalk Provider = "cloudtalk"

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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470762/media/cloudtalk_1722470761.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470762/media/cloudtalk_1722470761.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470762/media/cloudtalk_1722470761.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470762/media/cloudtalk_1722470761.svg",
			},
		},
	})
}
