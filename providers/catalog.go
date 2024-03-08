package providers

// ================================================================================
// Provider list
// ================================================================================

const (
	Salesforce        Provider = "salesforce"
	Hubspot           Provider = "hubspot"
	LinkedIn          Provider = "linkedIn"
	Salesloft         Provider = "salesloft"
	Outreach          Provider = "outreach"
	Pipedrive         Provider = "pipedrive"
	Sellsy            Provider = "sellsy"
	Attio             Provider = "attio"
	Close             Provider = "close"
	Keap              Provider = "keap"
	Asana             Provider = "asana"
	Dropbox           Provider = "dropbox"
	Notion            Provider = "notion"
	Gong              Provider = "gong"
	Zoom              Provider = "zoom"
	Intercom          Provider = "intercom"
	DocuSign          Provider = "docuSign"
	DocuSignDeveloper Provider = "docuSignDeveloper"
	Calendly          Provider = "calendly"
)

// ================================================================================
// Contains critical provider configuration (using types from types.gen.go)
// ================================================================================

var catalog = CatalogType{ // nolint:gochecknoglobals
	// Salesforce configuration
	Salesforce: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.my.salesforce.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://{{.workspace}}.my.salesforce.com/services/oauth2/authorize",
			TokenURL:                  "https://{{.workspace}}.my.salesforce.com/services/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "id",
				WorkspaceRefField: "instance_url",
				ScopesField:       "scope",
			},
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
			AuthURL:                   "https://app.hubspot.com/oauth/authorize",
			TokenURL:                  "https://api.hubapi.com/oauth/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
			AuthURL:                   "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL:                  "https://www.linkedin.com/oauth/v2/accessToken",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Salesloft configuration
	Salesloft: {
		AuthType: Oauth2,
		BaseURL:  "https://api.salesloft.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://accounts.salesloft.com/oauth/authorize",
			TokenURL:                  "https://accounts.salesloft.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Outreach configuration
	Outreach: {
		AuthType: Oauth2,
		BaseURL:  "https://api.outreach.io",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://api.outreach.io/oauth/authorize",
			TokenURL:                  "https://api.outreach.io/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Pipedrive configuration
	Pipedrive: {
		AuthType: Oauth2,
		BaseURL:  "https://api.pipedrive.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://oauth.pipedrive.com/oauth/authorize",
			TokenURL:                  "https://oauth.pipedrive.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Sellsy configuration
	Sellsy: {
		AuthType: Oauth2,
		BaseURL:  "https://api.sellsy.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://login.sellsy.com/oauth2/authorization",
			TokenURL:                  "https://login.sellsy.com/oauth2/access-tokens",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Attio configuration
	Attio: {
		AuthType: Oauth2,
		BaseURL:  "https://api.attio.com/api",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.attio.com/authorize",
			TokenURL:                  "https://app.attio.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Close configuration
	Close: {
		AuthType: Oauth2,
		BaseURL:  "https://api.close.com/api",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.close.com/oauth2/authorize",
			TokenURL:                  "https://api.close.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	Keap: {
		AuthType: Oauth2,
		BaseURL:  "https://api.infusionsoft.com",

		OauthOpts: OauthOpts{
			AuthURL:                   "https://accounts.infusionsoft.com/app/oauth/authorize",
			TokenURL:                  "https://api.infusionsoft.com/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Asana configuration
	Asana: {
		AuthType: Oauth2,
		BaseURL:  "https://app.asana.com/api",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.asana.com/-/oauth_authorize",
			TokenURL:                  "https://app.asana.com/-/oauth_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Dropbox configuration
	Dropbox: {
		AuthType: Oauth2,
		BaseURL:  "https://api.dropboxapi.com/2/",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://www.dropbox.com/oauth2/authorize",
			TokenURL:                  "https://api.dropboxapi.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Notion configuration
	Notion: {
		AuthType: Oauth2,
		BaseURL:  "https://api.notion.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://api.notion.com/v1/oauth/authorize",
			TokenURL:                  "https://api.notion.com/v1/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "owner.user.id",
				WorkspaceRefField: "workspace_id",
			},
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Gong configuration
	Gong: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.api.gong.io",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.gong.io/oauth2/authorize",
			TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Zoom configuration
	Zoom: {
		AuthType: Oauth2,
		BaseURL:  "https://api.zoom.us",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://zoom.us/oauth/authorize",
			TokenURL:                  "https://zoom.us/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Intercom configuration
	Intercom: {
		AuthType: Oauth2,
		BaseURL:  "https://api.intercom.io",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.intercom.com/oauth",
			TokenURL:                  "https://api.intercom.io/auth/eagle/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// DocuSign configuration
	DocuSign: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.docusign.net",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://account.docusign.com/oauth/auth",
			TokenURL:                  "https://account.docusign.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// DocuSign Developer configuration
	DocuSignDeveloper: {
		AuthType: Oauth2,
		BaseURL:  "https://demo.docusign.net",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://account-d.docusign.com/oauth/auth",
			TokenURL:                  "https://account-d.docusign.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Calendly configuration
	Calendly: {
		AuthType: Oauth2,
		BaseURL:  "https://api.calendly.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://auth.calendly.com/oauth/authorize",
			TokenURL:                  "https://auth.calendly.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Support: Support{
			BulkWrite: false,
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// SalesLoft configuration
	Salesloft: {
		AuthType: Oauth2,
		BaseURL:  "https://api.salesloft.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://accounts.salesloft.com/oauth/authorize",
			TokenURL:                  "https://accounts.salesloft.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
