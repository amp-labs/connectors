package providers

const Gong Provider = "gong"

func init() {
	// Gong configuration
	SetInfo(Gong, ProviderInfo{
		DisplayName: "Gong",
		AuthType:    Oauth2,
		// Gong API base URL is region-specific. The OAuth token response includes
		// an `api_base_url` field pointing to the tenant's regional endpoint.
		// US tenants get a tenant-specific URL, with a fallback to https://api.gong.io if missing.
		// EU/APAC tenants get a different URL, with no safe fallback.
		// Without this, non-US tenants receive "access token has been revoked" errors
		// because their token is valid only for their regional endpoint.
		BaseURL: "{{.api_base_url_for_customer}}",
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
				// See https://help.gong.io/docs/create-an-app-for-gong#exchange-the-code-for-an-access-token
				WorkspaceRefField: "api_base_url_for_customer",
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
