package providers

const Expensify Provider = "expensify"

func init() {
	SetInfo(Expensify, ProviderInfo{
		DisplayName: "Expensify",
		AuthType:    Custom,
		BaseURL:     "https://integrations.expensify.com/Integration-Server/ExpensifyIntegrations",
		CustomOpts: &CustomAuthOpts{
			Headers: []CustomAuthHeader{},
			Inputs: []CustomAuthInput{
				{
					Name:        "partnerUserID",
					DisplayName: "Partner User ID",
					DocsURL:     "https://integrations.expensify.com/Integration-Server/doc/#introduction",
				},
				{
					Name:        "partnerUserSecret",
					DisplayName: "Partner User Secret",
					DocsURL:     "https://integrations.expensify.com/Integration-Server/doc/#introduction",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771489872/media/expensify.com_1771489870.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771489902/media/expensify.com_1771489901.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771489872/media/expensify.com_1771489870.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1771489902/media/expensify.com_1771489901.svg",
			},
		},
	})
}
