package providers

const SendGrid Provider = "sendGrid"

func init() {
	// SendGrid configuration
	SetInfo(SendGrid, ProviderInfo{
		DisplayName: "SendGrid",
		AuthType:    ApiKey,
		BaseURL:     "https://api.sendgrid.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://www.twilio.com/docs/sendgrid",
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
