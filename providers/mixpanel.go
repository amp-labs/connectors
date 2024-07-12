package providers

const Mixpanel Provider = "mixpanel"

func init() {
	// Mixpanel configuration
	// workspace maps to either region(api, api-eu) or server(data,data-eu).
	// Supported Mixpanel APIs
	// -	Ingestion API
	// -	Identity API
	// -	Event Export API
	// -	Data Pipelines API

	SetInfo(Mixpanel, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.mixpanel.com",
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
