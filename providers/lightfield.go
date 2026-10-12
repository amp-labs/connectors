package providers

const Lightfield Provider = "lightfield"

func init() {
	SetInfo(Lightfield, ProviderInfo{
		DisplayName: "Lightfield",
		AuthType:    ApiKey,
		// API paths are versioned under /v1, e.g. https://api.lightfield.app/v1/accounts.
		// Every request also needs the `Lightfield-Version: 2026-03-01` header; the deep connector sets it.
		BaseURL: "https://api.lightfield.app",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.lightfield.app/using-the-api/api-keys",
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
