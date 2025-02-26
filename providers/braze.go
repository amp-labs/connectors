package providers

const Braze Provider = "braze"

func init() {
	SetInfo(Braze, ProviderInfo{
		DisplayName: "Braze",
		AuthType:    ApiKey,
		BaseURL:     "https://rest.{{.workspace}}.braze.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://www.braze.com/docs/developer_guide/home",
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
