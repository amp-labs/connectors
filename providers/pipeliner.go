package providers

const Pipeliner Provider = "pipeliner"

func init() {
	// Pipeliner API Key authentication
	SetInfo(Pipeliner, ProviderInfo{
		// TODO [ExplicitWorkspaceRequired: true]
		DisplayName: "Pipeliner",
		AuthType:    Basic,
		BaseURL:     "https://{{.region}}.api.pipelinersales.com/api/v100/rest/spaces/{{.workspace}}/entities",
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
