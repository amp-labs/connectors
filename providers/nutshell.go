package providers

const Nutshell Provider = "nutshell"

func init() {
	// Nutshell Configuration
	SetInfo(Nutshell, ProviderInfo{
		DisplayName: "Nutshell",
		AuthType: Basic,
		BaseURL:  "https://app.nutshell.com",
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
