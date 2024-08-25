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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722458409/media/amplitude_1722458408.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722458370/media/amplitude_1722458369.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722458435/media/amplitude_1722458435.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722458351/media/amplitude_1722458350.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: false,
	})
}
