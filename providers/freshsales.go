package providers

const Freshsales Provider = "freshsales"

func init() {
	SetInfo(Freshsales, ProviderInfo{
		DisplayName: "Freshsales",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.workspace}}.myfreshworks.com/crm/sales",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Token token=",
			},
			DocsURL: "https://developers.freshworks.com/crm/api/#authentication",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324573/media/freshsales_1722324572.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324555/media/freshsales_1722324554.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324573/media/freshsales_1722324572.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325347/media/freshsales_1722325345.svg",
			},
		},
	})
}
