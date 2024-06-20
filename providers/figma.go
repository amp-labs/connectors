package providers

const Figma Provider = "figma"

func init() {
	// Figma Support Configuration
	SetInfo(Figma, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.figma.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.figma.com/oauth",
			TokenURL:                  "https://www.figma.com/api/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user_id",
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
