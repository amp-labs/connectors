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
		PostAuthInfoNeeded: false,
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
