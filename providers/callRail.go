package providers

const CallRail Provider = "callRail"

func init() {
	// CallRail Configuration
	SetInfo(CallRail, ProviderInfo{
		DisplayName: "CallRail",
		AuthType: ApiKey,
		BaseURL:  "https://api.callrail.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Token token=",
			},
			DocsURL: "https://apidocs.callrail.com/#getting-started",
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
}
