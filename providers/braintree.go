package providers

const Braintree Provider = "braintree"

func init() {
	// Braintree Configuration
	SetInfo(Braintree, ProviderInfo{
		DisplayName: "Braintree",
		AuthType:    Basic,
		BaseURL:     "https://payments.sandbox.braintree-api.com/graphql",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722460381/media/braintree_1722460380.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722460341/media/braintree_1722460339.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722460381/media/braintree_1722460380.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722460360/media/braintree_1722460359.svg",
			},
		},
	})
}
