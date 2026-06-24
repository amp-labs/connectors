package providers

const (
	Square        Provider = "square"
	SquareSandbox Provider = "squareSandbox"
)

//nolint:funlen
func init() {
	SetInfo(Square, ProviderInfo{
		DisplayName: "Square",
		AuthType:    Oauth2,
		BaseURL:     "https://connect.squareup.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			AuthURL:   "https://connect.squareup.com/oauth2/authorize",
			AuthURLParams: map[string]string{
				"session": "true",
			},
			TokenURL:                  "https://connect.squareup.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://developer.squareup.com/docs/oauth-api/overview",
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "merchant_id",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270338/media/squareup.com_1782270337.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270357/media/squareup.com_1782270356.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270287/media/squareup.com_1782270286.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270322/media/squareup.com_1782270321.svg",
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

	SetInfo(SquareSandbox, ProviderInfo{
		DisplayName: "Square Sandbox",
		AuthType:    Oauth2,
		BaseURL:     "https://connect.squareupsandbox.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://connect.squareupsandbox.com/oauth2/authorize",
			TokenURL:                  "https://connect.squareupsandbox.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://developer.squareup.com/docs/devtools/sandbox/overview",
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "merchant_id",
			},
		},

		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270338/media/squareup.com_1782270337.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270357/media/squareup.com_1782270356.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270287/media/squareup.com_1782270286.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1782270322/media/squareup.com_1782270321.svg",
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
