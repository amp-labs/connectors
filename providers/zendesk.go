package providers

const (
	ZendeskChat                  Provider = "zendeskChat"
	ZendeskSunshineConversations Provider = "zendeskSunshineConversations"
	ZendeskSupport               Provider = "zendeskSupport"
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
					Name:        "workspace",
					DisplayName: "Subdomain",
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
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
				},
			},
		},
	})

	// Zendesk Sunshine Conversations configuration
	// Basic auth uses an API key ID as the username and its secret as the password.
	// Requests are made under /v2/apps/{appId}/..., so the app ID goes in the proxied path.
	// https://developer.zendesk.com/documentation/conversations/getting-started/api-authentication/
	SetInfo(ZendeskSunshineConversations, ProviderInfo{
		DisplayName: "Zendesk Sunshine Conversations",
		AuthType:    Basic,
		BaseURL:     "https://{{.workspace}}.zendesk.com/sc",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723143621/media/zendeskSunshineConversations_1723143620.png", // nolint:lll
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722102329/media/zendeskSupport_1722102328.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1723143621/media/zendeskSunshineConversations_1723143620.png", // nolint:lll
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
				},
			},
		},
	})
}
