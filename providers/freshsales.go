package providers

const Freshsales Provider = "freshsales"

func init() {
	SetInfo(Freshsales, ProviderInfo{
		DisplayName: "Freshsales",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.workspace}}.myfreshworks.com/crm/sales",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Token token=",
			},
			DocsURL: "https://developers.freshworks.com/crm/api/#authentication",
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
