package providers

// ================================================================================
// Provider list
// ================================================================================

const (
	Mock                                Provider = "mock"
	Salesforce                          Provider = "salesforce"
	Hubspot                             Provider = "hubspot"
	LinkedIn                            Provider = "linkedIn"
	Salesloft                           Provider = "salesloft"
	Outreach                            Provider = "outreach"
	Pipedrive                           Provider = "pipedrive"
	Copper                              Provider = "copper"
	ZohoCRM                             Provider = "zohoCRM"
	Klaviyo                             Provider = "klaviyo"
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
	MicrosoftDynamics365CRM             Provider = "microsoftDynamics365CRM"
	MicrosoftDynamics365BusinessCentral Provider = "microsoftDynamics365BusinessCentral"
	Gainsight                           Provider = "gainsight"
	Box                                 Provider = "box"
	ZendeskSupport                      Provider = "zendeskSupport"
	ZendeskChat                         Provider = "zendeskChat"
	WordPress                           Provider = "wordPress"
	Airtable                            Provider = "airtable"
	Slack                               Provider = "slack"
	HelpScoutMailbox                    Provider = "helpScoutMailbox"
	Timely                              Provider = "timely"
	Atlassian                           Provider = "atlassian"
	Webflow                             Provider = "webflow"
	Smartsheet                          Provider = "smartsheet"
	StackExchange                       Provider = "stackExchange"
	GoogleContacts                      Provider = "googleContacts"
	GoogleMail                          Provider = "googleMail"
)

// ================================================================================
// Contains critical provider configuration (using types from types.gen.go)
// ================================================================================

var catalog = CatalogType{ // nolint:gochecknoglobals
	Mock: {
		AuthType: None,
		BaseURL:  "https://not-a-real-domain.mock",
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		ProviderOpts: ProviderOpts{
			"isMock": "true",
		},
	},

	// Salesforce configuration
	Salesforce: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.my.salesforce.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
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
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: true,
				Delete: true,
			},
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
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.hubspot.com/oauth/authorize",
			TokenURL:                  "https://api.hubapi.com/oauth/v1/token",
			ExplicitScopesRequired:    true,
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
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL:                  "https://www.linkedin.com/oauth/v2/accessToken",
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
	},

	// Klaviyo configuration
	Klaviyo: {
		AuthType: Oauth2,
		BaseURL:  "https://a.klaviyo.com",
		OauthOpts: OauthOpts{
			GrantType:                 PKCE,
			AuthURL:                   "https://www.klaviyo.com/oauth/authorize",
			TokenURL:                  "https://a.klaviyo.com/oauth/token",
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
	},

	// Sellsy configuration
	Sellsy: {
		AuthType: Oauth2,
		BaseURL:  "https://api.sellsy.com",
		OauthOpts: OauthOpts{
			GrantType:                 PKCE,
			AuthURL:                   "https://login.sellsy.com/oauth2/authorization",
			TokenURL:                  "https://login.sellsy.com/oauth2/access-tokens",
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
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "data.id",
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
	},

	// Dropbox configuration
	Dropbox: {
		AuthType: Oauth2,
		BaseURL:  "https://api.dropboxapi.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://www.dropbox.com/oauth2/authorize",
			TokenURL:                  "https://api.dropboxapi.com/oauth2/token",
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
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "client_id",
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
	},

	// Calendly configuration
	Calendly: {
		AuthType: Oauth2,
		BaseURL:  "https://api.calendly.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.calendly.com/oauth/authorize",
			TokenURL:                  "https://auth.calendly.com/oauth/token",
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
	},

	MicrosoftDynamics365CRM: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.api.crm.dynamics.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
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
	},

	// Box configuration
	Box: {
		AuthType: Oauth2,
		BaseURL:  "https://api.box.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://account.box.com/api/oauth2/authorize",
			TokenURL:                  "https://api.box.com/oauth2/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Zendesk Support configuration
	ZendeskSupport: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.zendesk.com",
		OauthOpts: OauthOpts{
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	ZendeskChat: {
		AuthType: Oauth2,
		BaseURL:  "https://www.zopim.com",
		OauthOpts: OauthOpts{
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
	},

	// WordPress Support configuration
	WordPress: {
		AuthType: Oauth2,
		BaseURL:  "https://public-api.wordpress.com",
		OauthOpts: OauthOpts{
			AuthURL:                   "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL:                  "https://public-api.wordpress.com/oauth2/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Airtable Support Configuration
	Airtable: {
		AuthType: Oauth2,
		BaseURL:  "https://api.airtable.com",
		OauthOpts: OauthOpts{
			GrantType:                 PKCE,
			AuthURL:                   "https://airtable.com/oauth2/v1/authorize",
			TokenURL:                  "https://airtable.com/oauth2/v1/token",
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
	},

	// Slack configuration
	Slack: {
		AuthType: Oauth2,
		BaseURL:  "https://slack.com/api",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://slack.com/oauth/v2/authorize",
			TokenURL:                  "https://slack.com/api/oauth.v2.access",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:       "scope",
				WorkspaceRefField: "workspace_name",
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
	},
	// HelpScoutMailbox Support Configuration
	HelpScoutMailbox: {
		AuthType: Oauth2,
		BaseURL:  "https://api.helpscout.net",
		OauthOpts: OauthOpts{
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Timely Configuration
	Timely: {
		AuthType: Oauth2,
		BaseURL:  "https://api.timelyapp.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.timelyapp.com/1.1/oauth/authorize",
			TokenURL:                  "https://api.timelyapp.com/1.1/oauth/token",
			ExplicitScopesRequired:    false,
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
	},

	// Atlassian configuration
	Atlassian: {
		AuthType: Oauth2,
		BaseURL:  "https://api.atlassian.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.atlassian.com/authorize",
			TokenURL:                  "https://auth.atlassian.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
	},

	// Webflow Support Configuration
	Webflow: {
		AuthType: Oauth2,
		BaseURL:  "https://api.webflow.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://webflow.com/oauth/authorize",
			TokenURL:                  "https://api.webflow.com/oauth/access_token",
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
	},

	// Smartsheet Support Configuration
	Smartsheet: {
		AuthType: Oauth2,
		BaseURL:  "https://api.smartsheet.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.smartsheet.com/b/authorize",
			TokenURL:                  "https://api.smartsheet.com/2.0/token",
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
	},

	// StackExchange configuration
	StackExchange: {
		AuthType: Oauth2,
		BaseURL:  "https://api.stackexchange.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://stackoverflow.com/oauth",
			TokenURL:                  "https://stackoverflow.com/oauth/access_token/json",
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
	},

	// GoogleContacts Support Configuration
	GoogleContacts: {
		AuthType: Oauth2,
		BaseURL:  "https://people.googleapis.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:                  "https://oauth2.googleapis.com/token",
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
	},

	// GoogleMail Support Configuration
	GoogleMail: {
		AuthType: Oauth2,
		BaseURL:  "https://gmail.googleapis.com",
		OauthOpts: OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:                  "https://oauth2.googleapis.com/token",
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
	},
}
