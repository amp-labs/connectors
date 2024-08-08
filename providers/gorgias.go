package providers

const (
	Gorgias Provider = "gorgias"
)

func init() {
	// Gorgias Support Configuration
	SetInfo(Gorgias, ProviderInfo{
		DisplayName: "Gorgias",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.gorgias.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.gorgias.com/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.gorgias.com/oauth/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459392/media/gorgias_1722459391.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459336/media/gorgias_1722459335.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459373/media/gorgias_1722459372.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722459319/media/gorgias_1722459317.svg",
			},
		},
	})
}
