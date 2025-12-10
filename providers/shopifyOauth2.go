package providers

const ShopifyOAuth2 Provider = "shopify-oauth2"

func init() {
	SetInfo(ShopifyOAuth2, ProviderInfo{
		DisplayName: "Shopify",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.myshopify.com/admin/api/2025-10/graphql.json",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.myshopify.com/admin/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.myshopify.com/admin/oauth/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
