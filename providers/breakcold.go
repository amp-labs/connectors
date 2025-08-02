package providers

const Breakcold Provider = "breakcold"

func init() {
	SetInfo(Breakcold, ProviderInfo{
		DisplayName: "Breakcold",
		AuthType:    ApiKey,
		BaseURL:     "https://api.breakcold.com/rest",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-API-KEY",
			},
			DocsURL: "https://developer.breakcold.com/v3/api-reference/introduction#authentication",
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753365294/media/breakcold.com_1753365295.svg",
				LogoURL: "https://mintlify.s3.us-west-1.amazonaws.com/breakcold/logo/dark.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753365294/media/breakcold.com_1753365295.svg",
				LogoURL: "https://mintlify.s3.us-west-1.amazonaws.com/breakcold/logo/light.svg",
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
