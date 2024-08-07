package providers

const Hive Provider = "hive"

func init() {
	// Hive Connector Configuration
	SetInfo(Hive, ProviderInfo{
		DisplayName: "Hive",
		AuthType:    ApiKey,
		BaseURL:     "https://app.hive.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "api_key",
			},
			DocsURL: "https://developers.hive.com/reference/api-keys-and-auth",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410295/media/const%20Hive%20Provider%20%3D%20%22hive%22_1722410295.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410346/media/const%20Hive%20Provider%20%3D%20%22hive%22_1722410346.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410295/media/const%20Hive%20Provider%20%3D%20%22hive%22_1722410295.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722410346/media/const%20Hive%20Provider%20%3D%20%22hive%22_1722410346.svg",
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
