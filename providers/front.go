package providers

const Front Provider = "front"

func init() {
	SetInfo(Front, ProviderInfo{
		DisplayName: "Front",
		AuthType:    ApiKey,
		BaseURL:     "https://api2.frontapp.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://dev.frontapp.com/docs/create-and-revoke-api-tokens",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064519/media/front_1722064518.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064483/media/front_1722064482.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064547/media/front_1722064545.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064547/media/front_1722064545.svg",
			},
		},
	})
}
