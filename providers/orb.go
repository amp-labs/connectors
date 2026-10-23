package providers

const Orb Provider = "orb"

func init() {
	SetInfo(Orb, ProviderInfo{
		DisplayName: "Orb",
		AuthType:    ApiKey,
		// API paths are versioned under /v1, e.g. https://api.withorb.com/v1/customers.
		BaseURL: "https://api.withorb.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://docs.withorb.com/tutorials/first-usage-event",
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     false,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
