package providers

const ConstantContact Provider = "constantContact"

func init() {
	// ConstantContact configuration
	SetInfo(ConstantContact, ProviderInfo{
		DisplayName: "Constant Contact",
		AuthType:    Oauth2,
		BaseURL:     "https://api.cc.email",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165757/media/constantcontact.com_1722165755.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165757/media/constantcontact.com_1722165755.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165757/media/constantcontact.com_1722165755.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165757/media/constantcontact.com_1722165755.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://authz.constantcontact.com/oauth2/default/v1/authorize",
			TokenURL:                  "https://authz.constantcontact.com/oauth2/default/v1/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
