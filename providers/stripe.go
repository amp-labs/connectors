package providers

import "github.com/amp-labs/connectors/common"

const Stripe Provider = "stripe"

func init() {
	// Stripe configuration
	SetInfo(Stripe, ProviderInfo{
		DisplayName: "Stripe",
		AuthType:    ApiKey,
		BaseURL:     "https://api.stripe.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.stripe.com/keys",
		},
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://api.stripe.com/v1",
				DisplayName: "Stripe",
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
		PostAuthInfoNeeded: false,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456153/media/stripe_1722456152.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456095/media/stripe_1722456094.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456153/media/stripe_1722456152.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722456053/media/stripe_1722456051.svg",
			},
		},
	})
}
