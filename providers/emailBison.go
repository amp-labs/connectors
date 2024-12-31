package providers

const EmailBison Provider = "emailBison"

func init() {
	SetInfo(EmailBison, ProviderInfo{
		DisplayName: "EmailBison",
		AuthType:    ApiKey,
		BaseURL:     "https://dedi.emailbison.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://dedi.emailbison.com/api/reference",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://emailbison.com/apple-touch-icon.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734676433/media/emailbison.com_1734676433.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://emailbison.com/apple-touch-icon.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734676433/media/emailbison.com_1734676433.png",
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
