package providers

const ClariCopilot Provider = "clariCopilot"

func init() {
	SetInfo(ClariCopilot, ProviderInfo{
		DisplayName: "Clari Copliot",
		AuthType:    ApiKey,
		BaseURL:     "https://rest-api.copilot.clari.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "apiKey",
			},
			DocsURL: "https://api-doc.copilot.clari.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722337833/media/clari_1722337832.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722337810/media/clari_1722337809.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722337833/media/clari_1722337832.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722337781/media/clari_1722337779.svg",
			},
		},
	})
}
