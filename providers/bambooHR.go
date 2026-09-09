package providers

const BambooHR Provider = "bambooHR"

func init() {
	SetInfo(BambooHR, ProviderInfo{
		DisplayName: "BambooHR",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.bamboohr.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.bamboohr.com/authorize.php?request=authorize",
			TokenURL:                  "https://{{.workspace}}.bamboohr.com/token.php?request=token",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://www.bamboohr.com/favicon.ico",
				LogoURL: "https://www.bamboohr.com/favicon.ico",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://www.bamboohr.com/favicon.ico",
				LogoURL: "https://www.bamboohr.com/favicon.ico",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Company Domain",
					DocsURL:     "https://documentation.bamboohr.com/docs/getting-started",
					Prompt: "The subdomain in your BambooHR login URL. " +
						"For example, if you log in at https://mycompany.bamboohr.com, enter mycompany.",
				},
			},
		},
	})
}
