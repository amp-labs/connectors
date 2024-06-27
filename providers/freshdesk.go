package providers

const Freshdesk Provider = "freshdesk"

func init() {
	SetInfo(Freshdesk, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.freshdesk.com",
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
