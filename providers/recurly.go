package providers

const Recurly Provider = "recurly"

func init() {
	SetInfo(Recurly, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://v3.recurly.com",
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
