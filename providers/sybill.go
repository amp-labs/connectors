package providers

const Sybill Provider = "sybill"

func init() {
	SetInfo(Sybill, ProviderInfo{
		DisplayName: "Sybill",
		AuthType:    ApiKey,
		// API paths are versioned under /v1, e.g. https://api.sybill.ai/v1/conversations.
		// https://api.sybill.ai/docs/introduction.html#base-url
		BaseURL: "https://api.sybill.ai",
		// Organization API keys (sk_live_...) are sent as a Bearer token; reads need the `read` scope.
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://api.sybill.ai/docs/authentication.html",
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
