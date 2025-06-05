package providers

const KaseyaVSAX Provider = "kaseyaVSAX"

func init() {
	SetInfo(KaseyaVSAX, ProviderInfo{
		DisplayName: "Kaseya VSAX",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}",

		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749053435/menu-logo_mf2wiq.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749053435/menu-logo_mf2wiq.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749053597/logo-03615157_c5wzua.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749053597/logo-03615157_c5wzua.png",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Server Name",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
