package providers

const ClariCopilot Provider = "clariCopilot"

func init() {
	SetInfo(ClariCopilot, ProviderInfo{
		DisplayName: "Clari Copilot",
		AuthType:    Custom,
		BaseURL:     "https://rest-api.copilot.clari.com",
		CustomOpts: &CustomAuthOpts{
			Headers: []CustomAuthHeader{
				{
					Name:          "X-Api-Key",
					ValueTemplate: "{{ .apiKey }}",
				},
				{
					Name:          "X-Api-Password",
					ValueTemplate: "{{ .apiSecret }}",
				},
			},
			Inputs: []CustomAuthInput{
				{
					Name:        "apiKey",
					DisplayName: "API Key",
					DocsURL:     "https://api-doc.copilot.clari.com/",
					// Prompt:      "The API Key can be found in your Copilot workspace settings.",
				},
				{
					Name:        "apiSecret",
					DisplayName: "API Secret",
					DocsURL:     "https://api-doc.copilot.clari.com/",
					// Prompt:      "The API Secret can be found in your Copilot workspace settings, also called API password.",
				},
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name: "inputShouldNotBeCollectedForClariCopilot",
					ModuleDependencies: &ModuleDependencies{
						ModuleOtherModule: ModuleDependency{},
					},
				},
				{
					Name: "inputShouldBeCollectedForClariCopilot",
				},
			},
		},
	})
}
