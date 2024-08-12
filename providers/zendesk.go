package providers

const (
	ZendeskChat    Provider = "zendeskChat"
	ZendeskSupport Provider = "zendeskSupport"
)

func init() { // nolint:funlen
	// Zendesk Support configuration
	SetInfo(ZendeskSupport, ProviderInfo{
		DisplayName: "Zendesk Support",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.zendesk.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.zendesk.com/oauth/authorizations/new",
			TokenURL:                  "https://{{.workspace}}.zendesk.com/oauth/tokens",
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
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
	})

	// BLOCKED: refresh token seems to be one-time use.
	SetInfo(ZendeskChat, ProviderInfo{
		DisplayName: "Zendesk Chat",
		AuthType:    Oauth2,

		// Reference docs
		// https://developer.zendesk.com/documentation/live-chat/getting-started/auth/

		BaseURL: "https://{{.workspace}}.zendesk.com/api/v2/chat",
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
	})
}
