package providers

const Amplitude Provider = "amplitude"

func init() {
	// this connector supports the following apis
	/*
		Behavioral Cohorts
		CCPA DSAR
		Chart Annotations
		Dashboard REST
		Event Streaming Metrics Summary
		Export
		Releases*
		Taxonomy
		User Privacy
	*/
	SetInfo(Amplitude, ProviderInfo{
		DisplayName: "Amplitude",
		AuthType:    Basic,
		BaseURL:     "https://amplitude.com",
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
