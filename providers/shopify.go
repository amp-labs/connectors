package providers

const Shopify Provider = "shopify"

func init() {
	SetInfo(Shopify, ProviderInfo{
		DisplayName: "Shopify",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.workspace}}.myshopify.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Shopify-Access-Token",
			},
			DocsURL: "https://shopify.dev/docs/api/admin-rest#authentication",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326919/media/shopify_1722326918.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326847/media/shopify_1722326845.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326881/media/shopify_1722326880.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722326881/media/shopify_1722326880.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Store",
				},
			},
		},
	})
}
