package providers

const Rebilly Provider = "rebilly"

func init() {
	// Rebilly Configuration
	SetInfo(Rebilly, ProviderInfo{
		AuthType: ApiKey,
		BaseURL:  "https://api.rebilly.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "REB-APIKEY",
			},
			DocsURL: "https://www.rebilly.com/catalog/all/section/authentication/manage-api-keys",
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
