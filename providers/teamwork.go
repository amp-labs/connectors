package providers

const Teamwork Provider = "teamwork"

func init() {
	// Teamwork Configuration
	SetInfo(Teamwork, ProviderInfo{
		DisplayName: "Teamwork",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.teamwork.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.teamwork.com/launchpad/login",
			TokenURL:                  "https://www.teamwork.com/launchpad/v1/token.json",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user.id",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722404948/media/const%20Teamwork%20Provider%20%3D%20%22teamwork%22_1722404947.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722404979/media/const%20Teamwork%20Provider%20%3D%20%22teamwork%22_1722404979.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722404948/media/const%20Teamwork%20Provider%20%3D%20%22teamwork%22_1722404947.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722404979/media/const%20Teamwork%20Provider%20%3D%20%22teamwork%22_1722404979.svg",
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
