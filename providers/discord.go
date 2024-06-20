package providers

const Discord Provider = "discord"

func init() {
	// Discord Support Configuration
	SetInfo(Discord, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://discord.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://discord.com/oauth2/authorize",
			TokenURL:                  "https://discord.com/api/oauth2/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
