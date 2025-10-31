package providers

const Chargebee Provider = "chargebee"

func init() {
	// Chargebee configuration
	// workspace maps to site
	SetInfo(Chargebee, ProviderInfo{
		DisplayName: "Chargebee",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.chargebee.com/api",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326327/media/chargebee_1722326327.svg",
				// The logo may be not be observed in dark mode.
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326161/media/chargebee_1722326160.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326327/media/chargebee_1722326327.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326161/media/chargebee_1722326160.svg",
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
					DisplayName: "Site name",
				},
			},
		},
	})
}
