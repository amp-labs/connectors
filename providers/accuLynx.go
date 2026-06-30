package providers

const AccuLynx Provider = "accuLynx"

func init() {
	SetInfo(AccuLynx, ProviderInfo{
		DisplayName: "AccuLynx",
		AuthType:    ApiKey,
		BaseURL:     "https://api.acculynx.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://apidocs.acculynx.com/docs/authentication",
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
			Subscribe: true,
			Write:     true,
		},
		SubscribeRequirements: &SubscribeRequirements{
			// AccuLynx supports creating webhook subscriptions via API.
			// https://apidocs.acculynx.com/reference/postsubscription
			SubscribeByAPI: new(true),
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1777290175/media/acculynx.com_1777290175.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1777290119/media/acculynx.com_1777290118.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1777290175/media/acculynx.com_1777290175.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1777290119/media/acculynx.com_1777290118.svg",
			},
		},
	})
}
