package providers

const G2 Provider = "g2"

func init() {
	// G2 configuration
	SetInfo(G2, ProviderInfo{
		DisplayName: "G2",
		AuthType:    ApiKey,
		BaseURL:     "https://data.g2.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://data.g2.com/api/docs?shell#g2-v2-api",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722888708/media/data.g2.com_1722888706.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722888708/media/data.g2.com_1722888706.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722888708/media/data.g2.com_1722888706.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722888708/media/data.g2.com_1722888706.svg",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					DisplayName: "Product Id",
					Name:        "productId",
				},
			},
		},
	})
}
