package providers

const Hightouch Provider = "hightouch"

func init() {
	// Hightouch Connector Configuration
	SetInfo(Hightouch, ProviderInfo{
		DisplayName: "Hightouch",
		AuthType:    ApiKey,
		BaseURL:     "https://api.hightouch.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://hightouch.com/docs/api-reference#section/Authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765795916/Hightouch_Symbol_0_cqnb3w.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765796035/media/hightouch.com_1765796033.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765796003/Hightouch_Symbol_0_w4h02i.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1765796051/media/hightouch.com_1765796050.svg",
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
