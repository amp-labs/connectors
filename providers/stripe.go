package providers

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
		}, Support: Support{
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
