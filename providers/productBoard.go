package providers

const ProductBoard Provider = "productBoard"

func init() {
	// ProductBoard Configuration
	SetInfo(ProductBoard, ProviderInfo{
		DisplayName: "Product Board",
		AuthType:    Oauth2,
		BaseURL:     "https://api.productboard.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.productboard.com/oauth2/authorize",
			TokenURL:                  "https://app.productboard.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		///nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726225706/media/productboard.com/_1726225707.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726225776/media/productboard.com/_1726225777.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726225706/media/productboard.com/_1726225707.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1726225798/media/productboard.com/_1726225799.svg",
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
