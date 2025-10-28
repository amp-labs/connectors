package providers

const Xero Provider = "xero"

func init() {
	// Xero configuration
	SetInfo(Xero, ProviderInfo{
		DisplayName: "Xero",
		AuthType:    Oauth2,
		BaseURL:     "https://api.xero.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://login.xero.com/identity/connect/authorize",
			AuthURLParams:             map[string]string{"response_type": "code"},
			TokenURL:                  "https://identity.xero.com/connect/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		PostAuthInfoNeeded: true,
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     false,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1752052319/media/xero.com_1752052319.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1752052319/media/xero.com_1752052319.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1752052285/media/xero.com_1752052283.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1752052285/media/xero.com_1752052283.svg",
			},
		},

		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "tenantId",
				},
			},
		},
	})
}
