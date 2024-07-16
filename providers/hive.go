package providers

const Hive Provider = "hive"

func init() {
	// Hive Connector Configuration
	SetInfo(Hive, ProviderInfo{
		DisplayName: "Hive",
		AuthType:    ApiKey,
		BaseURL:     "https://app.hive.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "api_key",
			},
			DocsURL: "https://developers.hive.com/reference/api-keys-and-auth",
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
