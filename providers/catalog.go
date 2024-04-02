package providers

// ================================================================================
// Provider list
// ================================================================================

const (
	Salesforce                          Provider = "salesforce"
	Hubspot                             Provider = "hubspot"
	LinkedIn                            Provider = "linkedIn"
	Salesloft                           Provider = "salesloft"
	Outreach                            Provider = "outreach"
	Pipedrive                           Provider = "pipedrive"
	Copper                              Provider = "copper"
	ZohoCRM                             Provider = "zohoCRM"
	Sellsy                              Provider = "sellsy"
	Attio                               Provider = "attio"
	Close                               Provider = "close"
	Keap                                Provider = "keap"
	Asana                               Provider = "asana"
	Dropbox                             Provider = "dropbox"
	Notion                              Provider = "notion"
	Gong                                Provider = "gong"
	Zoom                                Provider = "zoom"
	Intercom                            Provider = "intercom"
	Capsule                             Provider = "capsule"
	DocuSign                            Provider = "docuSign"
	DocuSignDeveloper                   Provider = "docuSignDeveloper"
	Calendly                            Provider = "calendly"
	AWeber                              Provider = "aWeber"
	GetResponse                         Provider = "getResponse"
	ConstantContact                     Provider = "constantContact"
	MicrosoftDynamics365Sales           Provider = "microsoftDynamics365Sales"
	MicrosoftDynamics365BusinessCentral Provider = "microsoftDynamics365BusinessCentral"
	Gainsight                           Provider = "gainsight"
	GoogleCalendar                      Provider = "googleCalendar"
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

	// Capsule configuration
	Capsule: {
		AuthType: Oauth2,
		BaseURL:  "https://api.capsulecrm.com/api",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://api.capsulecrm.com/oauth/authorise",
			TokenURL:                  "https://api.capsulecrm.com/oauth/token",
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

	// Copper configuration
	Copper: {
		AuthType: Oauth2,
		BaseURL:  "https://api.copper.com/developer_api",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.copper.com/oauth/authorize",
			TokenURL:                  "https://app.copper.com/oauth/token",
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

	// ZohoCRM configuration
	ZohoCRM: {
		AuthType: Oauth2,
		BaseURL:  "https://www.zohoapis.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
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

	// GetResponse configuration
	GetResponse: {
		AuthType: Oauth2,
		BaseURL:  "https://api.getresponse.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://app.getresponse.com/oauth2_authorize.html",
			TokenURL:                  "https://api.getresponse.com/v3/token",
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

	// AWeber configuration
	AWeber: {
		AuthType: Oauth2,
		BaseURL:  "https://api.aweber.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://auth.aweber.com/oauth2/authorize",
			TokenURL:                  "https://auth.aweber.com/oauth2/token",
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

	// MS Sales configuration
	MicrosoftDynamics365Sales: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.api.crm.dynamics.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
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

	// ConstantContact configuration
	ConstantContact: {
		AuthType: Oauth2,
		BaseURL:  "https://api.cc.email",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://authz.constantcontact.com/oauth2/default/v1/authorize",
			TokenURL:                  "https://authz.constantcontact.com/oauth2/default/v1/token",
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

	// Microsoft Dynamics 365 Business Central configuration
	MicrosoftDynamics365BusinessCentral: {
		AuthType: Oauth2,
		BaseURL:  "https://api.businesscentral.dynamics.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
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

	// Gainsight configuration
	Gainsight: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.gainsightcloud.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://{{.workspace}}.gainsightcloud.com/v1/authorize",
			TokenURL:                  "https://{{.workspace}}.gainsightcloud.com/v1/users/oauth/token",
			ExplicitScopesRequired:    false,
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

	GoogleCalendar: {
		AuthType: Oauth2,
		BaseURL:  "https://www.googleapis.com/calendar",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:                  "https://oauth2.googleapis.com/token",
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
}
