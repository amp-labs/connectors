package providers

const (
	Livestorm Provider = "livestorm"
)

//nolint:funlen // init keeps a commented OAuth2 SetInfo next to the active API key registration for revert.
func init() {
	// OAuth2 (restore when client id/secret are available from Livestorm support):
	/*
		SetInfo(Livestorm, ProviderInfo{
			DisplayName: "Livestorm",
			AuthType:    Oauth2,
			BaseURL:     "https://api.livestorm.co",
			Oauth2Opts: &Oauth2Opts{
				GrantType:                 AuthorizationCode,
				AuthURL:                   "https://app.livestorm.co/oauth/authorize",
				TokenURL:                  "https://app.livestorm.co/oauth/token",
				ExplicitScopesRequired:    true,
				ExplicitWorkspaceRequired: false,
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
					IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158143/media/api.livestorm.co_1741158142.svg",
					LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158176/media/api.livestorm.co_1741158174.svg",
				},
				Regular: &MediaTypeRegular{
					IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158111/media/api.livestorm.co_1741158108.jpg",
					LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158160/media/api.livestorm.co_1741158158.svg",
				},
			},
		})
	*/

	// API token auth for development and integration tests until OAuth2 is restored above.
	// https://developers.livestorm.co/docs/api-token-authentication
	SetInfo(Livestorm, ProviderInfo{
		DisplayName: "Livestorm",
		AuthType:    ApiKey,
		BaseURL:     "https://api.livestorm.co",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
				// Livestorm sends the API token as the full Authorization header value (no Bearer prefix).
			},
			DocsURL: "https://developers.livestorm.co/docs/api-token-authentication",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158143/media/api.livestorm.co_1741158142.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158176/media/api.livestorm.co_1741158174.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158111/media/api.livestorm.co_1741158108.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741158160/media/api.livestorm.co_1741158158.svg",
			},
		},
	})
}
