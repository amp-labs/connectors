package providers

const Stripe Provider = "stripe"

func init() {
	// Stripe configuration
	SetInfo(Stripe, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.stripe.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:        InHeader, // Can also be InQuery
			HeaderName:  "Authorization",
			ValuePrefix: "Bearer ",
			DocsURL:     "https://api.6sense.com/docs/#get-your-api-token",
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
