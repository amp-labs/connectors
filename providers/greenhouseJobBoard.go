package providers

const GreenhouseJobBoard Provider = "greenhouseJobBoard"

//nolint:lll
func init() {
	// GreenHouseJobBoard Configuration
	SetInfo(GreenhouseJobBoard, ProviderInfo{
		DisplayName: "GreenhouseJobBoard",
		AuthType:    ApiKey,
		BaseURL:     "https://boards-api.greenhouse.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Basic ",
			},
			DocsURL: "https://developers.greenhouse.io",
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Delete: false,
				Insert: false,
				Update: false,
				Upsert: false,
			},
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760017955/media/developers.greenhouse.io_1760017960.jpg",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1760017999/media/developers.greenhouse.io_1760018005.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760017955/media/developers.greenhouse.io_1760017960.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760018070/media/developers.greenhouse.io_1760018076.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "board_token",
					DisplayName: "Board Token",
				},
			},
		},
	})
}
