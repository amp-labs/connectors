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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168426/media/sendgrid.com_1722168425.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168426/media/sendgrid.com_1722168425.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168426/media/sendgrid.com_1722168425.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722168426/media/sendgrid.com_1722168425.jpg",
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
