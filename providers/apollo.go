package providers

const Apollo = "apollo"

func init() {
	// Apollo API Key authentication
	SetInfo(Apollo, ProviderInfo{
		DisplayName: "Apollo",
		AuthType:    ApiKey,
		BaseURL:     "https://api.apollo.io",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Api-Key",
			},
			DocsURL: "https://apolloio.github.io/apollo-api-docs/?shell#authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410061/media/const%20Apollo%20%3D%20%22apollo%22_1722410061.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409884/media/const%20Apollo%20%3D%20%22apollo%22_1722409883.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410061/media/const%20Apollo%20%3D%20%22apollo%22_1722410061.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722409884/media/const%20Apollo%20%3D%20%22apollo%22_1722409883.svg",
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
