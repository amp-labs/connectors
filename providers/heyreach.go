package providers

const HeyReach Provider = "heyreach"

func init() {
	SetInfo(HeyReach, ProviderInfo{
		DisplayName: "heyreach",
		AuthType:    ApiKey,
		BaseURL:     "https://api.heyreach.io/api",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-API-KEY",
			},
			DocsURL: "https://documenter.getpostman.com/view/23808049/2sA2xb5F75",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735632745/media/heyreach.io_1735632744.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735632708/media/heyreach.io_1735632706.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735632745/media/heyreach.io_1735632744.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1735632708/media/heyreach.io_1735632706.png",
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
