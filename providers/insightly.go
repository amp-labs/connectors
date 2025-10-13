package providers

const Insightly = "insightly"

func init() {
	// Insightly API Key authentication
	SetInfo(Insightly, ProviderInfo{
		DisplayName: "Insightly",
		AuthType:    Basic,
		BaseURL:     "https://api.insightly.com",
		// Insightly expects the API key in the username field 
		// and the password field to be left blank.
		// See https://api.insightly.com/v3.1/Help#!
		BasicOpts: &BasicAuthOpts{
			ApiKeyAsBasic: true,
			ApiKeyAsBasicOpts: &ApiKeyAsBasicOpts{
				FieldUsed: UsernameField,
			},
			DocsURL: "https://support.insight.ly/en-US/Knowledge/article/2775#API%20Key%20and%20URL",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411056/media/const%20Insightly%20%3D%20%22insightly%22_1722411055.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411001/media/const%20Insightly%20%3D%20%22insightly%22_1722411000.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411056/media/const%20Insightly%20%3D%20%22insightly%22_1722411055.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722411001/media/const%20Insightly%20%3D%20%22insightly%22_1722411000.svg",
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
	})
}
