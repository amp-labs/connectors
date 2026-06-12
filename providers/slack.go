package providers

const Slack Provider = "slack"

func init() {
	// Slack configuration
	SetInfo(Slack, ProviderInfo{
		DisplayName: "Slack",
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
			Subscribe: false,
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
			Input: []MetadataItemInput{
				{
					DisplayName: "Signing Secret",
					DocsURL:     "https://docs.slack.dev/authentication/verifying-requests-from-slack/#validating-a-request",
					Name:        "signingSecret",
					Prompt:      "Grab your Slack 'Signing Secret', available in the app admin panel under Basic Info.",
				},
			},
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "teamId",
				},
			},
		},
	})
}
