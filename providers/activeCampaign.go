package providers

const ActiveCampaign Provider = "activeCampaign"

func init() {
	SetInfo(ActiveCampaign, ProviderInfo{
		DisplayName: "ActiveCampaign",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.workspace}}.api-us1.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Api-Token",
			},
			DocsURL: "https://developers.activecampaign.com/reference/authentication",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722880911/media/activeCampaign_1722880911.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722880869/media/activeCampaign_1722880869.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722880911/media/activeCampaign_1722880911.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722880896/media/activeCampaign_1722880896.svg",
			},
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
