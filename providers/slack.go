package providers

const (
	Slack          Provider = "slack"
	SlackUserScope Provider = "slackUserScope"
)

func init() { // nolint:funlen
	// Slack configuration
	SetInfo(Slack, ProviderInfo{
		DisplayName: "Slack Bot Scope",
		AuthType:    Oauth2,
		BaseURL:     "https://slack.com/api",
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
			Read:      true,
			Subscribe: true,
			Write:     true,
		},
		SubscribeRequirements: &SubscribeRequirements{
			SubscribeByAPI: new(false),
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225856/media/wo2jw59mssz2pk1eczur.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225856/media/wo2jw59mssz2pk1eczur.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722059419/media/slack_1722059417.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722059450/media/slack_1722059449.svg",
			},
		},
		PostAuthInfoNeeded: true,
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "teamId",
				},
			},
		},
	})

	SetInfo(SlackUserScope, ProviderInfo{
		DisplayName: "Slack User Scope",
		AuthType:    Oauth2,
		BaseURL:     "https://slack.com/api",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://slack.com/oauth/v2/authorize",
			TokenURL:                  "https://slack.com/api/oauth.v2.user.access",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			// https://docs.slack.dev/authentication/installing-with-oauth#how
			// > Bot scopes should be added as scope=<bot_scope>,
			// > and user scopes should be added as user_scope=<user_scope>.
			ScopeQueryParam: "user_scope",
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
		SubscribeRequirements: &SubscribeRequirements{
			SubscribeByAPI: new(false),
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225856/media/wo2jw59mssz2pk1eczur.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225856/media/wo2jw59mssz2pk1eczur.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722059419/media/slack_1722059417.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722059450/media/slack_1722059449.svg",
			},
		},
		PostAuthInfoNeeded: true,
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "teamId",
				},
			},
		},
	})
}
