package providers

const Reamaze Provider = "reamaze"

func init() {
	SetInfo(Reamaze, ProviderInfo{
		DisplayName: "Re:amaze",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.reamaze.io/api",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749811534/media/reamaze.com_1749811531.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749811534/media/reamaze.com_1749811531.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749811534/media/reamaze.com_1749811531.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749811534/media/reamaze.com_1749811531.png",
			},
		},
		PostAuthInfoNeeded: false,
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Account subdomain",
				},
			},
		},
	})
}
