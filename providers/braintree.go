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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
	})
}
