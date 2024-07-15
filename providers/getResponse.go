package providers

const GetResponse Provider = "getResponse"

func init() {
	// GetResponse configuration
	SetInfo(GetResponse, ProviderInfo{
		DisplayName: "GetResponse",
		AuthType:    Oauth2,
		BaseURL:     "https://api.getresponse.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.getresponse.com/oauth2_authorize.html",
			TokenURL:                  "https://api.getresponse.com/v3/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
