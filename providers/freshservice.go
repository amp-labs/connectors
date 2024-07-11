package providers

const Freshservice Provider = "freshservice"

func init() {
	SetInfo(Freshservice, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.freshservice.com",
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
