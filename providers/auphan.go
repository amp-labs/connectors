package providers

const (
	Auphan        Provider = "auphan"
	AuphanSandbox Provider = "auphanSandbox"
)

// nolint: funlen
func init() {
	// Auphan POS configuration
	SetInfo(Auphan, ProviderInfo{
		DisplayName: "Auphan",
		AuthType:    ApiKey,
		// The store ID is part of the hostname.
		BaseURL: "https://{{.workspace}}.gw1.auphansoftware.com/api/v1.php",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Auphan-Api-Token",
			},
			DocsURL: "https://api1.auphansoftware.com/resources/#authentication",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Store ID",
				},
				{
					Name:        "timezone",
					DisplayName: "Store Timezone",
					Prompt:      "IANA timezone of the store, e.g. America/Los_Angeles. Report dates are sent in this zone.",
				},
			},
		},

		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222398/Image_1794222393_0_hhgti6.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222291/Screenshot_1794222226_0_fvxwoj.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222398/Image_1794222393_0_hhgti6.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222291/Screenshot_1794222226_0_fvxwoj.png",
			},
		},
	})

	// Auphan POS sandbox configuration.
	// A hosted demo install that Auphan provisions per partner, with its own seeded store ID and token.
	SetInfo(AuphanSandbox, ProviderInfo{
		DisplayName: "Auphan Sandbox",
		AuthType:    ApiKey,
		BaseURL:     "https://api1.auphansoftware.com/{{.workspace}}/api/v1.php",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Auphan-Api-Token",
			},
			DocsURL: "https://api1.auphansoftware.com/resources/#authentication",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Demo Name",
				},
				{
					Name:        "storeId",
					DisplayName: "Store ID",
				},
				{
					Name:        "timezone",
					DisplayName: "Store Timezone",
					Prompt:      "IANA timezone of the store, e.g. America/Los_Angeles. Report dates are sent in this zone.",
				},
			},
		},

		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222398/Image_1794222393_0_hhgti6.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222291/Screenshot_1794222226_0_fvxwoj.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222398/Image_1794222393_0_hhgti6.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1794222291/Screenshot_1794222226_0_fvxwoj.png",
			},
		},
	})
}
