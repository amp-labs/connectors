package providers

// ================================================================================
// Provider list
// ================================================================================

const (
	Salesforce Provider = "salesforce"
	Hubspot    Provider = "hubspot"
	LinkedIn   Provider = "linkedIn"
)

// ================================================================================
// Contains critical provider configuration (using types from types.gen.go)
// ================================================================================

var Catalog = CatalogType{ // nolint:gochecknoglobals
	// Salesforce configuration
	Salesforce: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.my.salesforce.com",
		OauthOpts: OauthOpts{
			AuthURL:  "https://{{.workspace}}.my.salesforce.com/services/oauth2/authorize",
			TokenURL: "https://{{.workspace}}.my.salesforce.com/services/oauth2/token",
		},
		Support: Support{
			BulkWrite: true,
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		ProviderOpts: ProviderOpts{
			"restApiUrl": "https://{{.workspace}}.my.salesforce.com/services/data/v59.0",
			"domain":     "{{.workspace}}.my.salesforce.com",
		},
	},

	// Hubspot configuration
	Hubspot: {
		AuthType: Oauth2,
		BaseURL:  "https://api.hubapi.com",
		OauthOpts: OauthOpts{
			AuthURL:  "https://app.hubspot.com/oauth/authorize",
			TokenURL: "https://api.hubapi.com/oauth/v1/token",
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	},

	// LinkedIn configuration
	LinkedIn: {
		AuthType: Oauth2,
		BaseURL:  "https://api.linkedin.com",
		OauthOpts: OauthOpts{
			AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},
}
