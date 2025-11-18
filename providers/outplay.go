package providers

const Outplay Provider = "outplay"

func init() {
	SetInfo(Outplay, ProviderInfo{
		DisplayName: "Outplay",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}-api.outplayhq.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735777081/media/outplayhq.com_1735777080.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735777173/media/outplayhq.com_1735777172.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735777081/media/outplayhq.com_1735777080.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735777173/media/outplayhq.com_1735777172.png",
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
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Location",
				},
			},
		},
	})
}
