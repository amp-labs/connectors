package catalog

const HelpScoutMailbox Provider = "helpScoutMailbox"

func init() {
	// HelpScoutMailbox Support Configuration
	SetInfo(HelpScoutMailbox, ProviderInfo{
		DisplayName: "Help Scout Mailbox",
		AuthType:    Oauth2,
		BaseURL:     "https://api.helpscout.net",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://secure.helpscout.net/authentication/authorizeClientApplication",
			TokenURL:                  "https://api.helpscout.net/v2/oauth2/token",
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
