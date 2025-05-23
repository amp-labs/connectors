package providers

const Monday Provider = "monday"

func init() {
	// Monday Configuration
	SetInfo(Monday, ProviderInfo{
		DisplayName: "Monday",
		AuthType:    ApiKey,
		BaseURL:     "https://api.monday.com/",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: "header",
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
			},
			DocsURL: "https://developer.monday.com/api-reference/docs/authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722345745/media/monday_1722345745.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722345579/media/monday_1722345579.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722345745/media/monday_1722345745.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722345545/media/monday_1722345544.svg",
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
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
