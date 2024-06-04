package providers

// ================================================================================
// Provider list
// ================================================================================

const (
	Airtable                Provider = "airtable"
	AWeber                  Provider = "aWeber"
	Asana                   Provider = "asana"
	Atlassian               Provider = "atlassian"
	Attio                   Provider = "attio"
	Box                     Provider = "box"
	Calendly                Provider = "calendly"
	CampaignMonitor         Provider = "campaignMonitor"
	Capsule                 Provider = "capsule"
	ClickUp                 Provider = "clickup"
	Close                   Provider = "close"
	ConstantContact         Provider = "constantContact"
	Copper                  Provider = "copper"
	Discord                 Provider = "discord"
	Docusign                Provider = "docusign"
	DocusignDeveloper       Provider = "docusignDeveloper"
	Dropbox                 Provider = "dropbox"
	DropboxSign             Provider = "dropboxSign"
	Facebook                Provider = "facebook"
	Figma                   Provider = "figma"
	Gainsight               Provider = "gainsight"
	GetResponse             Provider = "getResponse"
	Gmail                   Provider = "gmail"
	Gong                    Provider = "gong"
	IroncladDemo            Provider = "ironcladDemo"
	IroncladEU              Provider = "ironcladEU"
	Ironclad                Provider = "ironclad"
	Google                  Provider = "google"
	GoogleContacts          Provider = "googleContacts"
	HelpScoutMailbox        Provider = "helpScoutMailbox"
	Hubspot                 Provider = "hubspot"
	Intercom                Provider = "intercom"
	Keap                    Provider = "keap"
	Klaviyo                 Provider = "klaviyo"
	LinkedIn                Provider = "linkedIn"
	DynamicsBusinessCentral Provider = "dynamicsBusinessCentral"
	DynamicsCRM             Provider = "dynamicsCRM"
	Miro                    Provider = "miro"
	Mock                    Provider = "mock"
	Monday                  Provider = "monday"
	Mural                   Provider = "mural"
	Notion                  Provider = "notion"
	Outreach                Provider = "outreach"
	Pinterest               Provider = "pinterest"
	Pipedrive               Provider = "pipedrive"
	RingCentral             Provider = "ringCentral"
	Salesforce              Provider = "salesforce"
	Salesloft               Provider = "salesloft"
	Sellsy                  Provider = "sellsy"
	ServiceNow              Provider = "serviceNow"
	Slack                   Provider = "slack"
	Smartsheet              Provider = "smartsheet"
	StackExchange           Provider = "stackExchange"
	TeamleaderCRM           Provider = "teamleaderCRM"
	Timely                  Provider = "timely"
	Typeform                Provider = "typeform"
	Webflow                 Provider = "webflow"
	WordPress               Provider = "wordPress"
	Wrike                   Provider = "wrike"
	ZendeskChat             Provider = "zendeskChat"
	ZendeskSupport          Provider = "zendeskSupport"
	ZohoCRM                 Provider = "zohoCRM"
	Zoom                    Provider = "zoom"
	Zuora                   Provider = "zuora"
	Aircall                 Provider = "aircall"
	Drift                   Provider = "drift"
	Microsoft               Provider = "microsoft"
	Formstack               Provider = "formstack"
	Aha                     Provider = "aha"
	SnapchatAds             Provider = "snapchatAds"
	Instagram               Provider = "instagram"
	Seismic                 Provider = "seismic"
	Github                  Provider = "github"
)

// ================================================================================
// Contains critical provider configuration (using types from types.gen.go)
// ================================================================================

var catalog = CatalogType{ // nolint:gochecknoglobals
	// Salesforce configuration
	Salesforce: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.my.salesforce.com",
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Outreach configuration
	Outreach: {
		AuthType: Oauth2,
		BaseURL:  "https://api.outreach.io",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		ProviderOpts: ProviderOpts{
			"restAPIURL": "https://api.outreach.io/api/v2",
		},
	},

	// RingCentral configuration

	RingCentral: {
		AuthType: Oauth2,
		BaseURL:  "https://platform.ringcentral.com",
		OauthOpts: &OauthOpts{
			GrantType:                 PKCE,
			AuthURL:                   "https://platform.ringcentral.com/restapi/oauth/authorize",
			TokenURL:                  "https://platform.ringcentral.com/restapi/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "owner_id",
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
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Capsule configuration
	Capsule: {
		AuthType: Oauth2,
		BaseURL:  "https://api.capsulecrm.com/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Wrikle configuration
	Wrike: {
		AuthType: Oauth2,
		BaseURL:  "https://www.wrike.com/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.wrike.com/oauth2/authorize",
			TokenURL:                  "https://www.wrike.com/oauth2/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Copper configuration
	Copper: {
		AuthType: Oauth2,
		BaseURL:  "https://api.copper.com/developer_api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// ZohoCRM configuration
	ZohoCRM: {
		DisplayName: "Zoho CRM",
		AuthType:    Oauth2,
		BaseURL:     "https://www.zohoapis.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "api_domain",
				ScopesField:       "scope",
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

	// Mural configuration
	Mural: {
		AuthType: Oauth2,
		BaseURL:  "https://api.mural.co/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.mural.co/oauth/authorize",
			TokenURL:                  "https://api.mural.co/oauth/token",
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

	// Klaviyo configuration
	Klaviyo: {
		AuthType: Oauth2,
		BaseURL:  "https://a.klaviyo.com",
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
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
		BaseURL:  "https://api.attio.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.attio.com/authorize",
			TokenURL:                  "https://app.attio.com/oauth/token",
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
	},

	// Close configuration
	Close: {
		AuthType: Oauth2,
		BaseURL:  "https://api.close.com/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.close.com/oauth2/authorize",
			TokenURL:                  "https://api.close.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "user_id",
				WorkspaceRefField: "organization_id",
				ScopesField:       "scope",
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

	Keap: {
		AuthType: Oauth2,
		BaseURL:  "https://api.infusionsoft.com",

		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Asana configuration
	Asana: {
		AuthType: Oauth2,
		BaseURL:  "https://app.asana.com/api",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Dropbox configuration
	Dropbox: {
		AuthType: Oauth2,
		BaseURL:  "https://api.dropboxapi.com",
		OauthOpts: &OauthOpts{
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
	},

	// Notion configuration
	Notion: {
		AuthType: Oauth2,
		BaseURL:  "https://api.notion.com",
		OauthOpts: &OauthOpts{
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
		BaseURL:  "https://api.gong.io",
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://app.gong.io/oauth2/authorize",
			TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Zoom configuration
	Zoom: {
		AuthType: Oauth2,
		BaseURL:  "https://api.zoom.us",
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://zoom.us/oauth/authorize",
			TokenURL:                  "https://zoom.us/oauth/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Intercom configuration
	Intercom: {
		AuthType: Oauth2,
		BaseURL:  "https://api.intercom.io",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	},

	// Docusign configuration
	Docusign: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.server}}.docusign.net",
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://account.docusign.com/oauth/auth",
			TokenURL:                  "https://account.docusign.com/oauth/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		PostAuthInfoNeeded: true,
	},

	// Docusign Developer configuration
	DocusignDeveloper: {
		AuthType: Oauth2,
		BaseURL:  "https://demo.docusign.net",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://account-d.docusign.com/oauth/auth",
			TokenURL:                  "https://account-d.docusign.com/oauth/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Calendly configuration
	Calendly: {
		AuthType: Oauth2,
		BaseURL:  "https://api.calendly.com",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// campaignMonitor configuration
	CampaignMonitor: {
		AuthType: Oauth2,
		BaseURL:  "https://api.createsend.com",
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://api.createsend.com/oauth",
			TokenURL:                  "https://api.createsend.com/oauth/token",
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

	// GetResponse configuration
	GetResponse: {
		AuthType: Oauth2,
		BaseURL:  "https://api.getresponse.com",
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	DynamicsCRM: {
		DisplayName: "Microsoft Dynamics CRM",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.api.crm.dynamics.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
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

	// ConstantContact configuration
	ConstantContact: {
		DisplayName: "Constant Contact",
		AuthType:    Oauth2,
		BaseURL:     "https://api.cc.email",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Microsoft Dynamics 365 Business Central configuration
	DynamicsBusinessCentral: {
		DisplayName: "Microsoft Dynamics Business Central",
		AuthType:    Oauth2,
		BaseURL:     "https://api.businesscentral.dynamics.com",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Gainsight configuration
	Gainsight: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.gainsightcloud.com",
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Zendesk Support configuration
	ZendeskSupport: {
		DisplayName: "Zendesk Support",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.zendesk.com",
		OauthOpts: &OauthOpts{
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
	},

	ZendeskChat: {
		DisplayName: "Zendesk Chat",
		AuthType:    Oauth2,
		BaseURL:     "https://www.zopim.com",
		OauthOpts: &OauthOpts{
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
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL:                  "https://public-api.wordpress.com/oauth2/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Airtable Support Configuration
	Airtable: {
		AuthType: Oauth2,
		BaseURL:  "https://api.airtable.com",
		OauthOpts: &OauthOpts{
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

	// Ironclad Support Configuration
	Ironclad: {
		AuthType: Oauth2,
		BaseURL:  "https://ironcladapp.com",
		OauthOpts: &OauthOpts{
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
	},

	// Slack configuration
	Slack: {
		AuthType: Oauth2,
		BaseURL:  "https://slack.com/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://slack.com/oauth/v2/authorize",
			TokenURL:                  "https://slack.com/api/oauth.v2.access",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},
	// HelpScoutMailbox Support Configuration
	HelpScoutMailbox: {
		DisplayName: "Help Scout Mailbox",
		AuthType:    Oauth2,
		BaseURL:     "https://api.helpscout.net",
		OauthOpts: &OauthOpts{
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
	},

	// Timely Configuration
	Timely: {
		AuthType: Oauth2,
		BaseURL:  "https://api.timelyapp.com",
		OauthOpts: &OauthOpts{
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
		DisplayName: "Atlassian Jira",
		AuthType:    Oauth2,
		BaseURL:     "https://api.atlassian.com",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Webflow Support Configuration
	Webflow: {
		AuthType: Oauth2,
		BaseURL:  "https://api.webflow.com",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Smartsheet Support Configuration
	Smartsheet: {
		AuthType: Oauth2,
		BaseURL:  "https://api.smartsheet.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.smartsheet.com/b/authorize",
			TokenURL:                  "https://api.smartsheet.com/2.0/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// StackExchange configuration
	StackExchange: {
		AuthType: Oauth2,
		BaseURL:  "https://api.stackexchange.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://stackoverflow.com/oauth",
			TokenURL:                  "https://stackoverflow.com/oauth/access_token/json",
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

	// Google Support Configuration
	Google: {
		AuthType: Oauth2,
		BaseURL:  "https://www.googleapis.com",
		OauthOpts: &OauthOpts{
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

	// GoogleContacts Support Configuration
	GoogleContacts: {
		DisplayName: "Google Contacts",
		AuthType:    Oauth2,
		BaseURL:     "https://people.googleapis.com",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// GoogleMail Support Configuration
	Gmail: {
		AuthType: Oauth2,
		BaseURL:  "https://gmail.googleapis.com",
		OauthOpts: &OauthOpts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	Monday: {
		AuthType: Oauth2,
		BaseURL:  "https://api.monday.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.monday.com/oauth2/authorize",
			TokenURL:                  "https://auth.monday.com/oauth2/token",
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
	// Figma Support Configuration
	Figma: {
		AuthType: Oauth2,
		BaseURL:  "https://api.figma.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.figma.com/oauth",
			TokenURL:                  "https://www.figma.com/api/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user_id",
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
	// Miro Support Configuration
	Miro: {
		AuthType: Oauth2,
		BaseURL:  "https://api.miro.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://miro.com/oauth/authorize",
			TokenURL:                  "https://api.miro.com/v1/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "user_id",
				WorkspaceRefField: "team_id",
				ScopesField:       "scope",
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
	Typeform: {
		AuthType: Oauth2,
		BaseURL:  "https://api.typeform.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.typeform.com/oauth/authorize",
			TokenURL:                  "https://api.typeform.com/oauth/token",
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

	// Zuora Configuration
	Zuora: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.zuora.com",
		OauthOpts: &OauthOpts{
			GrantType:                 ClientCredentials,
			AuthURL:                   "https://{{.workspace}}.zuora.com/oauth/auth_mock",
			TokenURL:                  "https://{{.workspace}}.zuora.com/oauth/token",
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

	// DropboxSign Configuration
	DropboxSign: {
		DisplayName: "Dropbox Sign",
		AuthType:    Oauth2,
		BaseURL:     "https://api.hellosign.com",
		OauthOpts: &OauthOpts{
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
	},

	// Facebook Ads Manager Configuration
	Facebook: {
		AuthType: Oauth2,
		BaseURL:  "https://graph.facebook.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.facebook.com/v19.0/dialog/oauth",
			TokenURL:                  "https://graph.facebook.com/v19.0/oauth/access_token",
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

	// ClickUp Support Configuration
	ClickUp: {
		AuthType: Oauth2,
		BaseURL:  "https://api.clickup.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.clickup.com/api",
			TokenURL:                  "https://api.clickup.com/api/v2/oauth/token",
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
	},

	// Discord Support Configuration
	Discord: {
		AuthType: Oauth2,
		BaseURL:  "https://discord.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://discord.com/oauth2/authorize",
			TokenURL:                  "https://discord.com/api/oauth2/token",
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

	// Drift Configuration
	Drift: {
		AuthType: Oauth2,
		BaseURL:  "https://driftapi.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://dev.drift.com/authorize",
			TokenURL:                  "https://driftapi.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "orgId",
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

	IroncladDemo: {
		AuthType: Oauth2,
		BaseURL:  "https://demo.ironcladapp.com",
		OauthOpts: &OauthOpts{
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
	},

	IroncladEU: {
		DisplayName: "Ironclad Europe",
		AuthType:    Oauth2,
		BaseURL:     "https://eu1.ironcladapp.com",
		OauthOpts: &OauthOpts{
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
	},

	// Aircall Configuration
	Aircall: {
		AuthType: Oauth2,
		BaseURL:  "https://api.aircall.io",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://dashboard.aircall.io/oauth/authorize",
			TokenURL:                  "https://api.aircall.io/v1/oauth/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Microsoft configuration
	Microsoft: {
		AuthType: Oauth2,
		BaseURL:  "https://graph.microsoft.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/common/oauth2/v2.0/token",
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Formstack configuration
	Formstack: {
		AuthType: Oauth2,
		BaseURL:  "https://www.formstack.com/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.formstack.com/api/v2/oauth2/authorize",
			TokenURL:                  "https://www.formstack.com/api/v2/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField: "user_id",
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

	// Pinterest configuration
	Pinterest: {
		AuthType: Oauth2,
		BaseURL:  "https://api.pinterest.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.pinterest.com/oauth",
			TokenURL:                  "https://api.pinterest.com/v5/oauth/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	Aha: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.aha.io/api",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.aha.io/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.aha.io/oauth/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Snapchat Ads configuration file
	SnapchatAds: {
		AuthType: Oauth2,
		BaseURL:  "https://adsapi.snapchat.com",
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://accounts.snapchat.com/login/oauth2/authorize",
			TokenURL:                  "https://accounts.snapchat.com/login/oauth2/access_token",
			ExplicitScopesRequired:    true,
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Instagram Configuration
	// TODO: Supports only short-lived tokens
	Instagram: {
		AuthType: Oauth2,
		BaseURL:  "https://graph.instagram.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://api.instagram.com/oauth/authorize",
			TokenURL:                  "https://api.instagram.com/oauth/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "user_id",
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

	// TeamleaderCRM Configuration
	TeamleaderCRM: {
		AuthType: Oauth2,
		BaseURL:  "https://api.focus.teamleader.eu",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://focus.teamleader.eu/oauth2/authorize",
			TokenURL:                  "https://focus.teamleader.eu/oauth2/access_token",
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

	// Github Configuration
	Github: {
		AuthType: Oauth2,
		BaseURL:  "https://api.github.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://github.com/login/oauth/authorize",
			TokenURL:                  "https://github.com/login/oauth/access_token",
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

	Seismic: {
		AuthType: Oauth2,
		BaseURL:  "https://api.seismic.com",
		OauthOpts: &OauthOpts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.seismic.com/tenants/{{.workspace}}/connect/authorize",
			TokenURL:                  "https://auth.seismic.com/tenants/{{.workspace}}/connect/token",
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

	// ServiceNow configuration
	ServiceNow: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.service-now.com",
		OauthOpts: &OauthOpts{
			AuthURL:                   "https://{{.workspace}}.service-now.com/oauth_auth.do",
			TokenURL:                  "https://{{.workspace}}.service-now.com/oauth_token.do",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
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
	},
}
