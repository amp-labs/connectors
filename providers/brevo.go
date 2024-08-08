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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325743/media/brevo_1722325742.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325614/media/brevo_1722325613.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325743/media/brevo_1722325742.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325555/media/brevo_1722325554.svg",
				// https://res.cloudinary.com/dycvts6vp/image/upload/v1722325684/media/brevo_1722325684.svg
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
