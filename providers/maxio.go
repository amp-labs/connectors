package providers

const Maxio Provider = "maxio"

func init() {
	SetInfo(Maxio, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.chargify.com",
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
