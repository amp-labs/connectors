package providers

const ShopifyApiKey Provider = "shopify-apikey"

func init() {
	// Shopify GraphQL Admin API Configuration
	//
	// IMPORTANT: Shopify OAuth2 Quirk
	// - Shopify uses OAuth2 to OBTAIN tokens (via AuthURL/TokenURL)
	// - However, tokens are "offline access tokens" that NEVER expire
	// - There is NO refresh token mechanism
	// - Once obtained, tokens act like permanent API keys

	// Docs: https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens
	SetInfo(ShopifyApiKey, ProviderInfo{
		DisplayName: "Shopify",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.workspace}}.myshopify.com/admin/api/2025-10/graphql.json",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "X-Shopify-Access-Token",
			},
			DocsURL: "https://shopify.dev/docs/api/admin-graphql#authentication",
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
					DocsURL:     "https://shopify.dev/docs/api/admin-graphql#endpoints",
				},
			},
		},
	})
}
