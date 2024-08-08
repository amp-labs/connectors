package providers

const (
	ZendeskChat     Provider = "zendeskChat"
	ZendeskSupport  Provider = "zendeskSupport"
	ZendeskSunshine Provider = "zendeskSunshineConversations"
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
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})

	SetInfo(ZendeskChat, ProviderInfo{
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

	SetInfo(ZendeskSunshine, ProviderInfo{
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
