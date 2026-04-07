package providers

const ConnectWise Provider = "connectWise"

func init() {
	SetInfo(ConnectWise, ProviderInfo{
		DisplayName: "ConnectWise",
		AuthType:    Basic,
		BaseURL:     "https://{{.region}}.myconnectwise.net",
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773780148/media/connectWise_1773780147.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773780188/media/connectWise_1773780188.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773780148/media/connectWise_1773780147.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1773780188/media/connectWise_1773780188.png",
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
			Input: []MetadataItemInput{{
				Name:         "region",
				DisplayName:  "Region",
				DefaultValue: "na",
				DocsURL:      "", // TODO link to real docs
			}},
		},
	})
}
