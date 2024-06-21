package providers

const (
	Dropbox     Provider = "dropbox"
	DropboxSign Provider = "dropboxSign"
)

func init() {
	// Dropbox configuration
	SetInfo(Dropbox, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://api.dropboxapi.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.dropbox.com/oauth2/authorize",
			TokenURL:                  "https://api.dropboxapi.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "account_id",
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

	SetInfo(DropboxSign, ProviderInfo{
		DisplayName: "Dropbox Sign",
		AuthType:    Oauth2,
		BaseURL:     "https://api.hellosign.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.hellosign.com/oauth/authorize",
			TokenURL:                  "https://app.hellosign.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
