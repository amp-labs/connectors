package providers

const ChargeOver Provider = "chargeOver"

func init() {
	SetInfo(ChargeOver, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://{{.workspace}}.chargeover.com",
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
