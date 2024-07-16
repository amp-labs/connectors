package providers

const Brevo Provider = "brevo"

func init() {
	// Brevo(Sendinblue) configuration
	SetInfo(Brevo, ProviderInfo{
		DisplayName: "Brevo",
		AuthType:    ApiKey,
		BaseURL:     "https://api.brevo.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "api-key",
			},
			DocsURL: "https://developers.brevo.com/docs/getting-started",
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
