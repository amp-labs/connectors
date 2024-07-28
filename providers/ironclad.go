package providers

const (
	Ironclad     Provider = "ironclad"
	IroncladDemo Provider = "ironcladDemo"
	IroncladEU   Provider = "ironcladEU"
)

func init() { //nolint:funlen
	// Ironclad Support Configuration
	SetInfo(Ironclad, ProviderInfo{
		DisplayName: "Ironclad",
		AuthType:    Oauth2,
		BaseURL:     "https://ironcladapp.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://ironcladapp.com/oauth/authorize",
			TokenURL:                  "https://ironcladapp.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
	})

	SetInfo(IroncladDemo, ProviderInfo{
		DisplayName: "Ironclad Demo",
		AuthType:    Oauth2,
		BaseURL:     "https://demo.ironcladapp.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://demo.ironcladapp.com/oauth/authorize",
			TokenURL:                  "https://demo.ironcladapp.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
	})

	SetInfo(IroncladEU, ProviderInfo{
		DisplayName: "Ironclad Europe",
		AuthType:    Oauth2,
		BaseURL:     "https://eu1.ironcladapp.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166230/media/ironcladapp.com_1722166229.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://eu1.ironcladapp.com/oauth/authorize",
			TokenURL:                  "https://eu1.ironcladapp.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
	})
}
