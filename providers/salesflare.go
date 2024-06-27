package providers

const Salesflare Provider = "salesflare"

func init() {
	// Salesflare configuration
	SetInfo(Salesflare, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.salesflare.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:        InHeader,
			HeaderName:  "Authorization",
			ValuePrefix: "Bearer ",
			DocsURL:     "https://api.salesflare.com/docs#section/Introduction/Authentication",
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
