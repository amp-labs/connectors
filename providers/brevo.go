package providers

const Brevo Provider = "brevo"

func init() {
	// Brevo(Sendinblue) configuration
	SetInfo(Brevo, ProviderInfo{
		DisplayName: "Brevo",
		AuthType:    ApiKey,
		BaseURL:     "https://api.brevo.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165255/media/brevo.com_1722165254.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165255/media/brevo.com_1722165254.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165255/media/brevo.com_1722165254.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722165255/media/brevo.com_1722165254.jpg",
			},
		},
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
