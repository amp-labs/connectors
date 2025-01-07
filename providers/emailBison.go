package providers

const EmailBison Provider = "emailBison"

func init() {
	SetInfo(EmailBison, ProviderInfo{
		DisplayName: "EmailBison",
		AuthType:    ApiKey,
		//Every Bison customer can have a full custom domain.
		BaseURL: "https://{{.workspace}}/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
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
