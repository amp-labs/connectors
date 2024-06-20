package providers

const Amplitude Provider = "amplitude"

func init() {
	// Amplitude Support Configuration
	SetInfo(Amplitude, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://amplitude.com",
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
