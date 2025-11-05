package providers

const GreenhouseJobBoard Provider = "greenhouseJobBoard"

//nolint:lll
func init() {
	// GreenHouseJobBoard Configuration
	SetInfo(GreenhouseJobBoard, ProviderInfo{
		DisplayName: "Greenhouse (Job board)",
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
			// This is tested & provider guide has been written, but we are not shipping it yet since
			// we haven't decided if this will be an independent connector.
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760017955/media/developers.greenhouse.io_1760017960.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760020366/media/developers.greenhouse.io_1760020371.svg",
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
					DocsURL:     "https://support.greenhouse.io/hc/en-us/articles/5888210160155-Find-your-job-board-URL",
				},
			},
		},
	})
}
