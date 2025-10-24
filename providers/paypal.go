package providers

import "github.com/amp-labs/connectors/common"

const (
	PayPal        Provider = "payPal"
	PayPalSandBox Provider = "payPalSandBox"
)

// nolint:funlen
func init() {
	// PayPal configuration file
	SetInfo(PayPal, ProviderInfo{
		DisplayName: "PayPal",
		AuthType:    Oauth2,
		BaseURL:     "https://api-m.paypal.com",
		Oauth2Opts: &Oauth2Opts{
			TokenURL:                  "https://api-m.paypal.com/v1/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 ClientCredentials,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DefaultModule: common.ModuleRoot,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967524/media/paypal.com_1760967523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967555/media/paypal.com_1760967555.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967524/media/paypal.com_1760967523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967578/media/paypal.com_1760967577.svg",
			},
		},
	})

	SetInfo(PayPalSandBox, ProviderInfo{
		DisplayName: "PayPal SandBox",
		AuthType:    Oauth2,
		BaseURL:     "https://api-m.sandbox.paypal.com",
		Oauth2Opts: &Oauth2Opts{
			TokenURL:                  "https://api-m.sandbox.paypal.com/v1/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 ClientCredentials,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DefaultModule: common.ModuleRoot,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967524/media/paypal.com_1760967523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967555/media/paypal.com_1760967555.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967524/media/paypal.com_1760967523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1760967578/media/paypal.com_1760967577.svg",
			},
		},
	})
}
