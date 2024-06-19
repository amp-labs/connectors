package providers

// ================================================================================
// Provider list
// ================================================================================

const (
	AcuityScheduling        Provider = "acuityScheduling"
	AdobeExperiencePlatform Provider = "adobeExperiencePlatform"
	Aha                     Provider = "aha"
	Aircall                 Provider = "aircall"
	Airtable                Provider = "airtable"
	Anthropic               Provider = "anthropic"
	Asana                   Provider = "asana"
	Atlassian               Provider = "atlassian"
	Attio                   Provider = "attio"
	AWeber                  Provider = "aWeber"
	Basecamp                Provider = "basecamp"
	BlueshiftEU             Provider = "blueshiftEU"
	Blueshift               Provider = "blueshift"
	Box                     Provider = "box"
	Calendly                Provider = "calendly"
	CampaignMonitor         Provider = "campaignMonitor"
	Capsule                 Provider = "capsule"
	ClickUp                 Provider = "clickup"
	Close                   Provider = "close"
	ConstantContact         Provider = "constantContact"
	Copper                  Provider = "copper"
	CustomerDataPipelines   Provider = "customerDataPipelines"
	CustomerJourneysTrack   Provider = "customerJourneysTrack"
	Discord                 Provider = "discord"
	Docusign                Provider = "docusign"
	DocusignDeveloper       Provider = "docusignDeveloper"
	Domo                    Provider = "domo"
	Drift                   Provider = "drift"
	Dropbox                 Provider = "dropbox"
	DropboxSign             Provider = "dropboxSign"
	DynamicsBusinessCentral Provider = "dynamicsBusinessCentral"
	DynamicsCRM             Provider = "dynamicsCRM"
	Facebook                Provider = "facebook"
	Figma                   Provider = "figma"
	Formstack               Provider = "formstack"
	Gainsight               Provider = "gainsight"
	GetResponse             Provider = "getResponse"
	Github                  Provider = "github"
	Gmail                   Provider = "gmail"
	Gong                    Provider = "gong"
	Google                  Provider = "google"
	GoogleContacts          Provider = "googleContacts"
	Guru                    Provider = "guru"
	HelpScoutMailbox        Provider = "helpScoutMailbox"
	Hubspot                 Provider = "hubspot"
	Hunter                  Provider = "hunter"
	Instagram               Provider = "instagram"
	Intercom                Provider = "intercom"
	Ironclad                Provider = "ironclad"
	IroncladDemo            Provider = "ironcladDemo"
	IroncladEU              Provider = "ironcladEU"
	Iterable                Provider = "iterable"
	Keap                    Provider = "keap"
	Klaviyo                 Provider = "klaviyo"
	LinkedIn                Provider = "linkedIn"
	Marketo                 Provider = "marketo"
	MessageBird             Provider = "messageBird"
	Microsoft               Provider = "microsoft"
	Miro                    Provider = "miro"
	Mixmax                  Provider = "mixmax"
	Mock                    Provider = "mock"
	Monday                  Provider = "monday"
	Mural                   Provider = "mural"
	Notion                  Provider = "notion"
	OpenAI                  Provider = "openAI"
	Outreach                Provider = "outreach"
	Pinterest               Provider = "pinterest"
	Pipedrive               Provider = "pipedrive"
	RingCentral             Provider = "ringCentral"
	Salesforce              Provider = "salesforce"
	Salesloft               Provider = "salesloft"
	Seismic                 Provider = "seismic"
	Sellsy                  Provider = "sellsy"
	ServiceNow              Provider = "serviceNow"
	SharePoint              Provider = "sharePoint"
	Slack                   Provider = "slack"
	Smartsheet              Provider = "smartsheet"
	SnapchatAds             Provider = "snapchatAds"
	StackExchange           Provider = "stackExchange"
	SurveyMonkey            Provider = "surveyMonkey"
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
)

// ================================================================================
// Contains critical provider configuration (using types from types.gen.go)
// ================================================================================

var catalog = CatalogType{ // nolint:gochecknoglobals
	// Salesforce configuration
	Salesforce: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.my.salesforce.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Salesloft configuration
	Salesloft: {
		AuthType: Oauth2,
		BaseURL:  "https://api.salesloft.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://accounts.salesloft.com/oauth/authorize",
			TokenURL:                  "https://accounts.salesloft.com/oauth/token",
			ExplicitScopesRequired:    false,
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.outreach.io/oauth/authorize",
			TokenURL:                  "https://api.outreach.io/oauth/token",
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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

	// Wrike configuration
	Wrike: {
		AuthType: Oauth2,
		BaseURL:  "https://www.wrike.com/api",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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

	CustomerDataPipelines: {
		AuthType: Basic,
		BaseURL:  "https://cdp.customer.io/v1",
		// DocsURL: https://customer.io/docs/api/cdp/#section/Authentication
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

	CustomerJourneysTrack: {
		AuthType: Basic,
		BaseURL:  "https://track.customer.io",
		// DocsURL: https://customer.io/docs/api/track/#section/Authentication
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
		DisplayName: "Zoho CRM",
		AuthType:    Oauth2,
		BaseURL:     "https://www.zohoapis.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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

		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.asana.com/-/oauth_authorize",
			TokenURL:                  "https://app.asana.com/-/oauth_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
	},

	// Notion configuration
	Notion: {
		AuthType: Oauth2,
		BaseURL:  "https://api.notion.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.notion.com/v1/oauth/authorize",
			TokenURL:                  "https://api.notion.com/v1/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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

	OpenAI: {
		AuthType: ApiKey,
		BaseURL:  "https://api.openai.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:        InHeader,
			HeaderName:  "Authorization",
			ValuePrefix: "Bearer ",
			DocsURL:     "https://platform.openai.com/docs/api-reference/api-keys",
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

	// Gong configuration
	Gong: {
		AuthType: Oauth2,
		BaseURL:  "https://api.gong.io",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.gong.io/oauth2/authorize",
			TokenURL:                  "https://app.gong.io/oauth2/generate-customer-token",
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://zoom.us/oauth/authorize",
			TokenURL:                  "https://zoom.us/oauth/token",
			ExplicitScopesRequired:    false,
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.intercom.com/oauth",
			TokenURL:                  "https://api.intercom.io/auth/eagle/token",
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
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	},

	// Docusign configuration
	Docusign: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.server}}.docusign.net",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://account.docusign.com/oauth/auth",
			TokenURL:                  "https://account.docusign.com/oauth/token",
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.createsend.com/oauth",
			TokenURL:                  "https://api.createsend.com/oauth/token",
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

	// GetResponse configuration
	GetResponse: {
		AuthType: Oauth2,
		BaseURL:  "https://api.getresponse.com",
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://auth.aweber.com/oauth2/authorize",
			TokenURL:                  "https://auth.aweber.com/oauth2/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	DynamicsCRM: {
		DisplayName: "Microsoft Dynamics CRM",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.api.crm.dynamics.com/api/data",
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:               "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			ExplicitScopesRequired: true,
			// TODO: flip this to false once we implement the ability to get workspace
			// information post-auth.
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
		DisplayName: "Constant Contact",
		AuthType:    Oauth2,
		BaseURL:     "https://api.cc.email",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/token",
			ExplicitScopesRequired:    true,
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
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://{{.workspace}}.gainsightcloud.com/v1/authorize",
			TokenURL:                  "https://{{.workspace}}.gainsightcloud.com/v1/users/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
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

	// Box configuration
	Box: {
		AuthType: Oauth2,
		BaseURL:  "https://api.box.com",
		Oauth2Opts: &Oauth2Opts{
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
	},

	ZendeskChat: {
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
	},

	// WordPress Support configuration
	WordPress: {
		AuthType: Oauth2,
		BaseURL:  "https://public-api.wordpress.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL:                  "https://public-api.wordpress.com/oauth2/token",
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
		Oauth2Opts: &Oauth2Opts{
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

	Anthropic: {
		AuthType: ApiKey,
		BaseURL:  "https://api.anthropic.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "x-api-key",
			DocsURL:    "https://docs.anthropic.com/en/api/getting-started#authentication",
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
	},

	// Timely Configuration
	Timely: {
		AuthType: Oauth2,
		BaseURL:  "https://api.timelyapp.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
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
	},

	// Facebook Ads Manager Configuration
	Facebook: {
		AuthType: Oauth2,
		BaseURL:  "https://graph.facebook.com",
		Oauth2Opts: &Oauth2Opts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// ClickUp Support Configuration
	ClickUp: {
		AuthType: Oauth2,
		BaseURL:  "https://api.clickup.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Pinterest configuration
	Pinterest: {
		AuthType: Oauth2,
		BaseURL:  "https://api.pinterest.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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

	SharePoint: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.sharepoint.com/_api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:               "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			ExplicitScopesRequired: true,
			// TODO: Switch to post-auth metadata collection
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

	// Instagram Configuration
	// TODO: Supports only short-lived tokens
	Instagram: {
		AuthType: Oauth2,
		BaseURL:  "https://graph.instagram.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://github.com/login/oauth/authorize",
			TokenURL:                  "https://github.com/login/oauth/access_token",
			ExplicitWorkspaceRequired: false,
			ExplicitScopesRequired:    true,
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

	Seismic: {
		AuthType: Oauth2,
		BaseURL:  "https://api.seismic.com",
		Oauth2Opts: &Oauth2Opts{
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
		Oauth2Opts: &Oauth2Opts{
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

	// AcuityScheduling Configuration
	AcuityScheduling: {
		AuthType: Oauth2,
		BaseURL:  "https://acuityscheduling.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://acuityscheduling.com/oauth2/authorize",
			TokenURL:                  "https://acuityscheduling.com/oauth2/token",
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

	// Basecamp Configuration
	// The wokspace in baseURL should be mapped to accounts.id
	Basecamp: {
		AuthType: Oauth2,
		BaseURL:  "https://3.basecampapi.com/{{.workspace}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://launchpad.37signals.com/authorization/new?type=web_server",
			TokenURL:                  "https://launchpad.37signals.com/authorization/token?type=refresh",
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

	// SurveyMonkey configuration file
	SurveyMonkey: {
		AuthType: Oauth2,
		BaseURL:  "https://api.surveymonkey.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.surveymonkey.com/oauth/authorize",
			TokenURL:                  "https://api.surveymonkey.com/oauth/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	},

	// Domo configuration file
	Domo: {
		AuthType: Oauth2,
		BaseURL:  "https://api.domo.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://api.domo.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "userId",
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

	// Iterable API Key authentication
	Iterable: {
		AuthType: ApiKey,
		BaseURL:  "https://api.iterable.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "Api-Key",
			DocsURL:    "https://app.iterable.com/settings/apiKeys",
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

	// Marketo configuration file
	// workspace maps to marketo instance
	Marketo: {
		AuthType: Oauth2,
		BaseURL:  "https://{{.workspace}}.mktorest.com",
		Oauth2Opts: &Oauth2Opts{
			TokenURL:                  "https://{{.workspace}}.mktorest.com/identity/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 ClientCredentials,
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

	// Hunter Connector Configuration
	Hunter: {
		AuthType: ApiKey,
		BaseURL:  "https://api.hunter.io/",
		ApiKeyOpts: &ApiKeyOpts{
			Type:           InQuery,
			QueryParamName: "api_key",
			DocsURL:        "https://hunter.io/api-documentation#authentication",
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

	// Guru API Key authentication
	Guru: {
		AuthType: ApiKey,
		BaseURL:  "https://api.getguru.com",
		ApiKeyOpts: &ApiKeyOpts{
			Type:       InHeader,
			HeaderName: "Api-Key",
			DocsURL:    "https://developer.getguru.com/docs/getting-started",
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

	// AdobeExperiencePlatform 2-legged auth
	AdobeExperiencePlatform: {
		AuthType: Oauth2,
		BaseURL:  "https://platform.adobe.io",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://ims-na1.adobelogin.com/ims/token/v3",
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
}
