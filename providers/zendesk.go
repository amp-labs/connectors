package providers

import "github.com/amp-labs/connectors/common"

const (
	ZendeskChat    Provider = "zendeskChat"
	ZendeskSupport Provider = "zendeskSupport"
)

const (
	// ModuleZendeskTicketing is used for proxying requests through.
	// https://developer.zendesk.com/api-reference/ticketing/introduction/
	ModuleZendeskTicketing common.ModuleID = "ticketing"
	// ModuleZendeskHelpCenter is Zendesk Help Center.
	// https://developer.zendesk.com/api-reference/help_center/help-center-api/introduction/
	ModuleZendeskHelpCenter common.ModuleID = "help-center"
)

func init() { // nolint:funlen
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
		Modules: &Modules{
			ModuleZendeskTicketing: {
				BaseURL:     "https://{{.workspace}}.zendesk.com/api/v2",
				DisplayName: "Zendesk Ticketing",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleZendeskHelpCenter: {
				BaseURL:     "https://{{.workspace}}.zendesk.com/api/v2",
				DisplayName: "Zendesk Help Center",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/wkaellrizizwvelbdl6r.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102329/media/zendeskSupport_1722102328.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724364159/media/tmk9w2cxvmfxrms9qwjq.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102362/media/zendeskSupport_1722102361.svg",
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name: "workspace",
				},
			},
		},
	})

	// BLOCKED: refresh token seems to be one-time use.
	SetInfo(ZendeskChat, ProviderInfo{
		DisplayName: "Zendesk Chat",
		AuthType:    Oauth2,

		// Reference docs
		// https://developer.zendesk.com/documentation/live-chat/getting-started/auth/

		BaseURL: "https://{{.workspace}}.zendesk.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.zendesk.com/oauth2/chat/authorizations/new",
			TokenURL:                  "https://{{.workspace}}.zendesk.com/oauth2/chat/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102329/media/zendeskSupport_1722102328.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102329/media/zendeskSupport_1722102328.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102362/media/zendeskSupport_1722102361.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102362/media/zendeskSupport_1722102361.svg",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name: "workspace",
				},
			},
		},
	})
}
