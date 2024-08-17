package providers

const Adobe Provider = "adobe"

//nolint:lll
func init() {
	// Adobe 2-legged auth
	SetInfo(Adobe, ProviderInfo{
		DisplayName: "Adobe",
		AuthType:    Oauth2,
		BaseURL:     "https://platform.adobe.io",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://ims-na1.adobelogin.com/ims/token/v3",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722065581/media/adobeExperiencePlatform_1722065579.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722065536/media/adobeExperiencePlatform_1722065535.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722065619/media/adobeExperiencePlatform_1722065617.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722065555/media/adobeExperiencePlatform_1722065554.svg",
			},
		},
	})
}
