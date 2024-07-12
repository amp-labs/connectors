package providers

const (
	BlueshiftEU Provider = "blueshiftEU"
	Blueshift   Provider = "blueshift"
)

func init() {
	// Blueshift configuration
	SetInfo(Blueshift, ProviderInfo{
		DisplayName: "Blueshift",
		AuthType:    Basic,
		BaseURL:     "https://api.getblueshift.com/api",
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

	// BlueshiftEU connfiguration
	SetInfo(BlueshiftEU, ProviderInfo{
		DisplayName: "Blueshift (EU)",
		AuthType:    Basic,
		BaseURL:     "https://api.eu.getblueshift.com/api",
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
