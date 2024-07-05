package providers

const Pipeliner Provider = "pipeliner"

func init() {
	// Pipeliner API Key authentication
	SetInfo(Pipeliner, ProviderInfo{
		// TODO [ExplicitWorkspaceRequired: true]
		DisplayName: "Pipeliner",
		AuthType:    Basic,
		BaseURL:     "https://eu-central.api.pipelinersales.com",
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
