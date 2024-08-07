package providers

const Coda Provider = "coda"

func init() {
	// Coda Configuration
	SetInfo(Coda, ProviderInfo{
		DisplayName: "Coda",
		AuthType:    ApiKey,
		BaseURL:     "https://coda.io/apis",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://coda.io/developers/apis/v1#section/Introduction",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459966/media/coda_1722459965.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459898/media/coda_1722459896.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459941/media/coda_1722459941.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459917/media/coda_1722459916.svg",
			},
		},
	})
}
