package providers

const CallRail Provider = "callRail"

func init() {
	// CallRail Configuration
	SetInfo(CallRail, ProviderInfo{
		DisplayName: "CallRail",
		AuthType:    ApiKey,
		BaseURL:     "https://api.callrail.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Token token=",
			},
			DocsURL: "https://apidocs.callrail.com/#getting-started",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722461906/media/callRail_1722461906.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722461886/media/callRail_1722461886.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722461906/media/callRail_1722461906.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722461856/media/callRail_1722461853.svg",
			},
		},
	})
}
