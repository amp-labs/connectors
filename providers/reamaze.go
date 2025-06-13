package providers

const Reamaze Provider = "reamaze"

func init() {
	SetInfo(Reamaze, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.reamaze.io/api",
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
