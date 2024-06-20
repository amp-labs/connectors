package providers

const Brevo Provider = "brevo"

func init() {
	apiKeyOpts := &ApiKeyOpts{
		Type: InHeader,
	}

	if err := apiKeyOpts.MergeApiKeyInHeaderOpts(ApiKeyInHeaderOpts{
		HeaderName: "api-key",
		DocsURL:    "https://developers.brevo.com/docs/getting-started",
	}); err != nil {
		panic(err)
	}

	// Brevo(Sendinblue) configuration
	SetInfo(Brevo, ProviderInfo{
		AuthType:   ApiKey,
		BaseURL:    "https://api.brevo.com",
		ApiKeyOpts: apiKeyOpts,
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
