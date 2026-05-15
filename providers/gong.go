package providers

const Gong Provider = "gong"

func init() {
	// Gong configuration
	SetInfo(Gong, ProviderInfo{
		DisplayName: "Gong",
		AuthType:    Oauth2,
		// Gong API base URL is region-specific. The OAuth token response includes
		// an `api_base_url` field pointing to the tenant's regional endpoint.
		// US tenants get https://api.gong.io; EU/APAC tenants get a different URL.
		// Without this, non-US tenants receive "access token has been revoked" errors
		// because their token is valid only for their regional endpoint.
		BaseURL: "https://{{.api_base_url}}",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327371/media/gong_1722327370.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327434/media/gong_1722327433.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327392/media/gong_1722327391.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327416/media/gong_1722327415.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.gong.io/oauth2/authorize",
			TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "api_base_url",
				ScopesField:       "scope",
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
			Subscribe: true,
			Write:     true,
		},
		SubscribeRequirements: &SubscribeRequirements{
			SubscribeByAPI: new(false),
		},
	})
}
