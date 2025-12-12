package providers

const (
	Adyen     Provider = "adyen"
	AdyenTest Provider = "adyenTest"
)

func init() { // nolint:funlen
	// Ayden configuration
	SetInfo(Adyen, ProviderInfo{
		DisplayName: "Adyen",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.adyen.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://ca-live.adyen.com/ca/ca/oauth/connect.shtml",
			TokenURL:                  "https://oauth-live.adyen.com/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760952588/media/adyen.com_1760952587.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760710249/media/adyen.com_1760710247.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760952588/media/adyen.com_1760952587.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760710278/media/adyen.com_1760710278.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
				},
			},
		},
	})

	// Ayden configuration
	SetInfo(AdyenTest, ProviderInfo{
		DisplayName: "Adyen Test",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.adyen.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCodePKCE,
			AuthURL:                   "https://ca-test.adyen.com/ca/ca/oauth/connect.shtml",
			TokenURL:                  "https://oauth-test.adyen.com/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760952588/media/adyen.com_1760952587.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760710249/media/adyen.com_1760710247.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760952588/media/adyen.com_1760952587.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760710278/media/adyen.com_1760710278.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
				},
			},
		},
	})
}
