package providers

const Bynder Provider = "bynder"

func init() {
	SetInfo(Bynder, ProviderInfo{
		DisplayName: "Bynder",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.bynder.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.bynder.com/v6/authentication/oauth2/auth",
			TokenURL:                  "https://{{.workspace}}.bynder.com/v6/authentication/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329798/media/bynder_1722329797.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329763/media/bynder_1722329761.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724363929/media/wqzogvbxncn0hj6qpvfp.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329821/media/bynder_1722329820.svg",
			},
		},
	})
}
