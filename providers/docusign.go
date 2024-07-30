package providers

const (
	Docusign          Provider = "docusign"
	DocusignDeveloper Provider = "docusignDeveloper"
)

func init() {
	// Docusign configuration
	SetInfo(Docusign, ProviderInfo{
		DisplayName: "Docusign",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.server}}.docusign.net",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://account.docusign.com/oauth/auth",
			TokenURL:                  "https://account.docusign.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320728/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320727.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320768/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320768.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320728/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320727.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320864/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320863.svg",
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
		PostAuthInfoNeeded: true,
	})

	// Docusign Developer configuration
	SetInfo(DocusignDeveloper, ProviderInfo{
		DisplayName: "Docusign Developer",
		AuthType:    Oauth2,
		BaseURL:     "https://demo.docusign.net",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://account-d.docusign.com/oauth/auth",
			TokenURL:                  "https://account-d.docusign.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320728/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320727.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320768/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320768.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320728/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320727.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722320864/media/Docusign%20%20%20%20%20%20%20%20%20%20Provider%20%3D%20%22docusign%22_1722320863.svg",
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
