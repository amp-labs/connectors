package providers

const (
	Ontraport Provider = "ontraport"
)

func init() { //nolint:funlen
	SetInfo(Ontraport, ProviderInfo{
		DisplayName: "Ontraport",
		AuthType:    ApiKey,
		BaseURL:     "https://api.ontraport.com/1/",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Api-Key",
			},
			DocsURL: "https://api.ontraport.com/doc/#obtain-an-api-key",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733863601/media/ontraport_1733863600.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733863622/media/ontraport_1733863622.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733863601/media/ontraport_1733863600.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733863636/media/ontraport_1733863636.svg",
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
