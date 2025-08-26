package providers

const Loxo Provider = "loxo"

//nolint:lll
func init() {
	// Loxo configuration
	SetInfo(Loxo, ProviderInfo{
		DisplayName: "Loxo",
		AuthType:    ApiKey,
		BaseURL:     "https://pod4.app.loxo.co/api/integration-user-loxo-withampersand-com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://help.loxo.co/en/articles/446640-loxo-s-open-api#h_668904ffbd",
		}, Support: Support{
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
				IconURL: "https://static.intercomassets.com/avatars/988166/square_128/custom_avatar-1664218549.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1756216233/media/loxo.co_1756216238.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://static.intercomassets.com/avatars/988166/square_128/custom_avatar-1664218549.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1756216216/media/loxo.co_1756216221.png",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "domain",
					DisplayName: "Full Domain",
				},
				{
					Name:        "agency_slug",
					DisplayName: "Agency Slug",
				},
			},
		},
	})
}
