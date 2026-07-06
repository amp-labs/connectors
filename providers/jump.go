package providers

const Jump Provider = "jump"

func init() {
	SetInfo(Jump, ProviderInfo{
		DisplayName: "Jump",
		AuthType:    ApiKey,
		BaseURL:     "https://my.jumpapp.com/enterprise/graphql",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://my.jumpapp.com/enterprise/documentation/authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782351055/media/jumpapp.com_1782351055.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782351024/media/jumpapp.com_1782351023.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782351055/media/jumpapp.com_1782351055.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782351024/media/jumpapp.com_1782351023.svg",
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
