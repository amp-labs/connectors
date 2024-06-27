package providers

const (
	GladlyDev  Provider = "gladlyDev"
	GladlyProd Provider = "gladlyProd"
)

func init() {
	// GladlyDev configuration
	SetInfo(GladlyDev, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.gladly.qa",
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

	// GladlyProd configuration
	SetInfo(GladlyProd, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.gladly.com",
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
