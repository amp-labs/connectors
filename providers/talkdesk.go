package providers

const Talkdesk Provider = "talkdesk"

func init() {
	SetInfo(Talkdesk, ProviderInfo{
		DisplayName: "Talkdesk",
		AuthType:    Oauth2,
		// Talkdesk deploys in multiple geographic locations
		// US - api.talkdeskapp.com
		// Europe - api.talkdeskapp.eu
		// Canada - api.talkdeskappca.com
		// Australia - api.mytalkdesk.au
		// UK - api.talkdeskapp.co.uk
		BaseURL: "https://{{.talkdesk_api_domain}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			// US - talkdeskid.com
			// EU - talkdeskid.eu
			// AU - talkdeskid.au
			// CA - talkdeskidca.com
			AuthURL:                   "https://{{.workspace}}.{{.talkdesk_token_domain}}/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.{{.talkdesk_token_domain}}/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733431983/media/talkdesk.com_1733431982.png",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1733432333/media/talkdesk.com_1733432332.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733431983/media/talkdesk.com_1733431982.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733432426/media/talkdesk.com_1733432426.svg",
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
					DisplayName: "Account name",
				},
				{
					Name:         "talkdesk_api_domain",
					DefaultValue: "api.talkdeskapp.com",
					DisplayName:  "Talkdesk API Domain",
					Prompt:       "Provide Your Regional API domain: US(api.talkdeskapp.com), EU(api.talkdeskapp.eu)...",
					DocsURL:      "https://docs.talkdesk.com/docs/how-to-guarantee-your-app-works-in-all-regions#supported-regions-and-base-urls", // nolint:lll
				},
				{
					Name:         "talkdesk_token_domain",
					DefaultValue: "talkdeskid.com",
					DisplayName:  "Talkdesk Token Domain",
					Prompt:       "Provide Your Regional API Token domain: US(talkdeskid.com), EU(talkdeskid.eu)...",
					DocsURL:      "https://docs.talkdesk.com/reference/authorization-code-basic-post",
				},
			},
		},
	})
}
