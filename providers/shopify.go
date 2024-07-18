package providers

const Shopify Provider = "shopify"

func init() {
	SetInfo(Shopify, ProviderInfo{
		DisplayName: "Shopify",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.workspace}}.myshopify.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Shopify-Access-Token",
			},
			DocsURL: "https://shopify.dev/docs/api/admin-rest#authentication",
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
