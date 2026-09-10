package providers

const BambooHR Provider = "bambooHR"

func init() {
	SetInfo(BambooHR, ProviderInfo{
		DisplayName: "BambooHR",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.bamboohr.com/api",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1788955884/media/bamboohr.com_1788955882.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1788955924/media/bamboohr.com_1788955923.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1788955884/media/bamboohr.com_1788955882.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1788955924/media/bamboohr.com_1788955923.svg",
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
