package providers

const (
	RingOverEU Provider = "ringOverEU"
	RingOverUS Provider = "ringOverUS"
)

//nolint:funlen
func init() {
	SetInfo(RingOverUS, ProviderInfo{
		DisplayName: "RingOver (US)",
		AuthType:    ApiKey,
		BaseURL:     "https://public-api-us.ringover.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
			},
			DocsURL: "https://developer.ringover.com",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1740678708/media/app.ringover.com_1740678707.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740678624/media/app.ringover.com_1740678624.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740678502/ringOver_vkuk42.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740678649/media/app.ringover.com_1740678648.svg",
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

	SetInfo(RingOverEU, ProviderInfo{
		DisplayName: "RingOver (EU)",
		AuthType:    ApiKey,
		BaseURL:     "https://public-api.ringover.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
			},
			DocsURL: "https://developer.ringover.com",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1740678708/media/app.ringover.com_1740678707.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740678624/media/app.ringover.com_1740678624.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740678502/ringOver_vkuk42.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740678649/media/app.ringover.com_1740678648.svg",
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
