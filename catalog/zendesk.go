package catalog

const (
	ZendeskChat    Provider = "zendeskChat"
	ZendeskSupport Provider = "zendeskSupport"
)

func init() {
	// Zendesk Support configuration
	SetInfo(ZendeskSupport, ProviderInfo{
		DisplayName: "Zendesk Support",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.zendesk.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.zendesk.com/oauth/authorizations/new",
			TokenURL:                  "https://{{.workspace}}.zendesk.com/oauth/tokens",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
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

	SetInfo(ZendeskChat, ProviderInfo{
		DisplayName: "Zendesk Chat",
		AuthType:    Oauth2,
		BaseURL:     "https://www.zopim.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.zopim.com/oauth2/authorizations/new?subdomain={{.workspace}}",
			TokenURL:                  "https://www.zopim.com/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
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
