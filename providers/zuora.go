package providers

const Zuora Provider = "zuora"

func init() {
	// Zuora Configuration
	SetInfo(Zuora, ProviderInfo{
		DisplayName: "Zuora",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.zuora.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://{{.workspace}}.zuora.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722063502/media/zuora_1722063501.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722063345/media/zuora_1722063343.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722063502/media/zuora_1722063501.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722063469/media/zuora_1722063468.svg",
			},
		},
	})
}
