package providers

const Mailgun Provider = "mailgun"

func init() {
	SetInfo(Mailgun, ProviderInfo{
		DisplayName: "Mailgun",
		AuthType:    Basic,
		BaseURL:     "https://api.mailgun.net",
		BasicOpts: &BasicAuthOpts{
			DocsURL: "https://documentation.mailgun.com/docs/mailgun/api-reference/authentication/",
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					// The Mailgun sending domain. Required for domain-scoped
					// objects (bounces, complaints, templates, etc.); substituted
					// into the URL path at read time.
					Name:        "workspace",
					DisplayName: "Domain name",
					DocsURL:     "https://documentation.mailgun.com/docs/mailgun/api-reference/intro/",
				},
				{
					// Resolved to the regional base URL in the connector.
					Name:         "region",
					DisplayName:  "Region",
					DefaultValue: "us",
					Prompt:       `Mailgun API region: "us" (api.mailgun.net) or "eu" (api.eu.mailgun.net).`,
					DocsURL:      "https://documentation.mailgun.com/docs/mailgun/api-reference/intro/",
				},
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy: true,
			// Read stays false until the post-merge integration-testing
			// enablement PR flips it (with screenshots), per the standard flow.
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071456/media/mailgun_1722071455.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071433/media/mailgun_1722071431.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071495/media/mailgun_1722071493.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722071474/media/mailgun_1722071473.svg",
			},
		},
	})
}
