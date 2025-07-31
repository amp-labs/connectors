package providers

const QuickBooks Provider = "quickbooks"

func init() {
	SetInfo(QuickBooks, ProviderInfo{
		DisplayName: "QuickBooks",
		AuthType:    Oauth2,
		BaseURL:     "https://sandbox-quickbooks.api.intuit.com",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753941999/media/quickbooks.com_1753941998.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753942143/media/quickbooks.com_1753942142.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753941999/media/quickbooks.com_1753941998.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753942143/media/quickbooks.com_1753942142.svg",
			},
		},
	})
}
