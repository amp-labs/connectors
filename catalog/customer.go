package catalog

const (
	CustomerDataPipelines Provider = "customerDataPipelines"
	CustomerJourneysApp   Provider = "customerJourneysApp"
	CustomerJourneysTrack Provider = "customerJourneysTrack"
)

func init() {
	SetInfo(CustomerDataPipelines, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://cdp.customer.io/v1",
		// DocsURL: https://customer.io/docs/api/cdp/#section/Authentication
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
	})

	SetInfo(CustomerJourneysApp, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.customer.io",
		ApiKeyOpts: &ApiKeyOpts{
			Type:        InHeader,
			HeaderName:  "Authorization",
			ValuePrefix: "Bearer ",
			DocsURL:     "https://customer.io/docs/api/app/#section/Authentication",
		},
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
	})

	SetInfo(CustomerJourneysTrack, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://track.customer.io",
		// DocsURL: https://customer.io/docs/api/track/#section/Authentication
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
	})
}
