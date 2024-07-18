package providers

const Zoom Provider = "zoom"

func init() {
	// Zoom configuration
	SetInfo(Zoom, ProviderInfo{
		DisplayName: "Zoom",
		AuthType:    Oauth2,
		BaseURL:     "https://api.zoom.us",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://zoom.us/oauth/authorize",
			TokenURL:                  "https://zoom.us/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
	})
}
