package providers

const ChartMogul Provider = "chartMogul"

func init() {
	SetInfo(ChartMogul, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://api.chartmogul.com",
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
		PostAuthInfoNeeded: true,
	})
}
