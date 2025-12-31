package providers

const Segment Provider = "segment"

func init() {
	// Segment configuration
	SetInfo(Segment, ProviderInfo{
		DisplayName: "Segment",
		AuthType:    ApiKey,
		BaseURL:     "https://api.segmentapis.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.segmentapis.com/tag/Getting-Started#section/Get-an-API-token",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766478745/media/segment.com_1766478744.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766478830/media/segment.com_1766478830.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766478732/Symbol_wxl8ii.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1766478773/media/segment.com_1766478772.svg",
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
