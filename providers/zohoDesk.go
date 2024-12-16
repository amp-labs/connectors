package providers

const ZohoDesk Provider = "zohoDesk"

func init() {
	SetInfo(ZohoDesk, ProviderInfo{
		DisplayName: "Zoho Desk",
		AuthType:    Oauth2,
		BaseURL:     "https://desk.zoho.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			AuthURL:   "https://accounts.zoho.com/oauth/v2/auth",
			AuthURLParams: map[string]string{
				"access_type": "offline",
			},
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734171152/zohodeskIcon_qp6nv3.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734171557/zohodeskLogoRegular_u6akdl.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734171152/zohodeskIcon_qp6nv3.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734171446/zohodeskLogoRegular_kuxqpz.svg",
			},
		},
	})
}
