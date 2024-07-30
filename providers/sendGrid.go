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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330743/media/sendGrid_1722330741.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330795/media/sendGrid_1722330795.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330743/media/sendGrid_1722330741.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722330818/media/sendGrid_1722330817.svg",
			},
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
