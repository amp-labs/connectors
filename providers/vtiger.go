package providers

const Vtiger Provider = "vtiger"

func init() {
	SetInfo(Vtiger, ProviderInfo{
		DisplayName: "Vtiger",
		AuthType:    Basic,
		// Every vtiger customer can have a full custom domain.
		BaseURL: "https://{{.workspace}}/restapi",
		BasicOpts: &BasicAuthOpts{
			DocsURL: "https://help.vtiger.com/faq/140159403-What-is-Access-Key",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749627816/media/vtiger.com_1749627825.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749627849/media/vtiger.com_1749627859.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749627816/media/vtiger.com_1749627825.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1749627849/media/vtiger.com_1749627859.png",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Full Domain",
				},
			},
		},
	})
}
