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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328016/media/ironclad_1722328015.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327987/media/ironclad_1722327987.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328016/media/ironclad_1722328015.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327967/media/ironclad_1722327967.svg",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})

	SetInfo(IroncladDemo, ProviderInfo{
		DisplayName: "Ironclad Demo",
		AuthType:    Oauth2,
		BaseURL:     "https://demo.ironcladapp.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328016/media/ironclad_1722328015.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327987/media/ironclad_1722327987.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328016/media/ironclad_1722328015.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327967/media/ironclad_1722327967.svg",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})

	SetInfo(IroncladEU, ProviderInfo{
		DisplayName: "Ironclad Europe",
		AuthType:    Oauth2,
		BaseURL:     "https://eu1.ironcladapp.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328016/media/ironclad_1722328015.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327987/media/ironclad_1722327987.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328016/media/ironclad_1722328015.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327967/media/ironclad_1722327967.svg",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Labels: &Labels{
			LabelExperimental: LabelValueTrue,
		},
	})
}
