package providers

import "github.com/amp-labs/connectors/common"

const ConstantContact Provider = "constantContact"

func init() {
	// ConstantContact configuration
	SetInfo(ConstantContact, ProviderInfo{
		DisplayName: "Constant Contact",
		AuthType:    Oauth2,
		BaseURL:     "https://api.cc.email",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326707/media/constantContact_1722326706.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326746/media/constantContact_1722326745.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326707/media/constantContact_1722326706.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326772/media/constantContact_1722326771.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://authz.constantcontact.com/oauth2/default/v1/authorize",
			TokenURL:                  "https://authz.constantcontact.com/oauth2/default/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://api.cc.email/v3",
				DisplayName: "Constant Contact",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
