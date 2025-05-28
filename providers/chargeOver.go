package providers

const ChargeOver Provider = "chargeOver"

func init() {
	SetInfo(ChargeOver, ProviderInfo{
		DisplayName: "ChargeOver",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.chargeover.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722460983/media/chargeover_1722460983.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722461005/media/chargeover_1722461004.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722460983/media/chargeover_1722460983.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722461005/media/chargeover_1722461004.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Domain name",
					DocsURL:     "https://developer.chargeover.com/docs/api-information/getting-started/",
				},
			},
		},
	})
}
