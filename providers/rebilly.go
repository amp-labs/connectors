package providers

const Rebilly Provider = "rebilly"

func init() {
	// Rebilly Configuration
	SetInfo(Rebilly, ProviderInfo{
		DisplayName: "Rebilly",
		AuthType:    ApiKey,
		BaseURL:     "https://api.rebilly.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "REB-APIKEY",
			},
			DocsURL: "https://www.rebilly.com/catalog/all/section/authentication/manage-api-keys",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724224447/media/noypybveuwpupubnizyo.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722408423/media/const%20Rebilly%20Provider%20%3D%20%22rebilly%22_1722408423.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722408171/media/const%20Rebilly%20Provider%20%3D%20%22rebilly%22_1722408170.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722408423/media/const%20Rebilly%20Provider%20%3D%20%22rebilly%22_1722408423.svg",
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
	})
}
