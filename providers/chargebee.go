package providers

const Chargebee Provider = "chargebee"

func init() {
	// Chargebee connfiguration
	// workspace maps to site
	SetInfo(Chargebee, ProviderInfo{
		DisplayName: "Chargebee",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.chargebee.com/api",
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
