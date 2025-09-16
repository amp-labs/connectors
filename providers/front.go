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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225074/media/xx1wx03acobxieiddokq.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225074/media/xx1wx03acobxieiddokq.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064547/media/front_1722064545.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064547/media/front_1722064545.svg",
			},
		},
	})
}
