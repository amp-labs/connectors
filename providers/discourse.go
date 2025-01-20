package providers

const Discourse Provider = "discourse"

func init() {
	SetInfo(Discourse, ProviderInfo{
		DisplayName: "Discourse",
		AuthType:    ApiKey,
		// Discourse is self-hosted, and the domain on which it is hosted serves as the base URL.
		BaseURL: "https://{{.workspace}}",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Api-Key",
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

		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734557159/media/discourse.org_1734557159.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734557186/media/discourse.org_1734557186.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1734557116/media/discourse.org_1734557115.jpg",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1734557138/media/discourse.org_1734557138.svg",
			},
		},
	})
}
