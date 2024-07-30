package providers

const Figma Provider = "figma"

func init() {
	// Figma Support Configuration
	SetInfo(Figma, ProviderInfo{
		DisplayName: "Figma",
		AuthType:    Oauth2,
		BaseURL:     "https://api.figma.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.figma.com/oauth",
			TokenURL:                  "https://www.figma.com/api/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user_id",
			},
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722323536/media/const%20Figma%20Provider%20%3D%20%22figma%22_1722323535.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722323505/media/const%20Figma%20Provider%20%3D%20%22figma%22_1722323505.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722323536/media/const%20Figma%20Provider%20%3D%20%22figma%22_1722323535.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722323344/media/const%20Figma%20Provider%20%3D%20%22figma%22_1722323344.svg",
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
