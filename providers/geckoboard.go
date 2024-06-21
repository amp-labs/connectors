package providers

const Geckoboard Provider = "geckoboard"

func init() {
	SetInfo(Geckoboard, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://api.geckoboard.com",
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
		PostAuthInfoNeeded: true,
	})
}
