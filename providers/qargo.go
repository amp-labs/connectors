package providers

const Qargo Provider = "qargo"

func init() {
	SetInfo(Qargo, ProviderInfo{
		DisplayName: "Qargo",
		AuthType:    Oauth2,
		// API paths are versioned under /v1, e.g. https://api.qargo.com/v1/companies/company.
		BaseURL: "https://api.qargo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType: ClientCredentials,
			// The token endpoint takes the client id and secret as Basic auth and allows
			// 5 requests per hour per tenant, so tokens must be reused until they expire.
			TokenURL:                  "https://api.qargo.com/v1/auth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://api-docs.qargo.com/docs/section/authentication",
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
