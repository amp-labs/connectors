package providers

const Flatfile Provider = "flatfile"

func init() {
	SetInfo(Flatfile, ProviderInfo{
		DisplayName: "Flatfile",
		AuthType:    ApiKey,
		BaseURL:     "https://api.x.flatfile.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://reference.flatfile.com/overview/welcome",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1750081977/media/flatfile.com_1750081977.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1750082071/media/flatfile.com_1750082071.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1750100764/flatfileicon_gjerpl_ag7via.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1750082112/media/flatfile.com_1750082111.svg",
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})
}
