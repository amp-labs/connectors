package providers

const (
	Ironclad     Provider = "ironclad"
	IroncladDemo Provider = "ironcladDemo"
	IroncladEU   Provider = "ironcladEU"
)

func init() { //nolint:funlen
	// Ironclad Support Configuration
	SetInfo(Ironclad, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://ironcladapp.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://ironcladapp.com/oauth/authorize",
			TokenURL:                  "https://ironcladapp.com/oauth/token",
			ExplicitScopesRequired:    true,
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})

	SetInfo(IroncladDemo, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://demo.ironcladapp.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://demo.ironcladapp.com/oauth/authorize",
			TokenURL:                  "https://demo.ironcladapp.com/oauth/token",
			ExplicitScopesRequired:    true,
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})

	SetInfo(IroncladEU, ProviderInfo{
		DisplayName: "Ironclad Europe",
		AuthType:    Oauth2,
		BaseURL:     "https://eu1.ironcladapp.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://eu1.ironcladapp.com/oauth/authorize",
			TokenURL:                  "https://eu1.ironcladapp.com/oauth/token",
			ExplicitScopesRequired:    true,
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
