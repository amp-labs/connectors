package providers

const Stripe Provider = "stripe"

func init() {
	// Stripe configuration
	SetInfo(Stripe, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.stripe.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:        InHeader,
			HeaderName:  "Authorization",
			ValuePrefix: "Bearer ",
			DocsURL:     "https://docs.stripe.com/keys",
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
	})
}
