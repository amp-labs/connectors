package providers

const Gladly Provider = "gladly"

func init() {
	// Gladly configuration
	SetInfo(Gladly, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.gladly.qa",
		// this is my organization/workspace = partner-withampersand.us-uat
		// for live production "https://{{organization}}.gladly.com",
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
