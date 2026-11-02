package providers

const SugarCRM Provider = "sugarCRM"

func init() {
	// SugarCRM renamed itself SugarAI in 2026. The provider name stays sugarCRM so that
	// existing references keep working; only the display name follows the rebrand.
	//
	// Sugar authenticates with the OAuth 2.0 password grant only, against a token endpoint
	// that also takes a "platform" field (logins without one count as "base", the web
	// app). Sugar allows one session per user per platform, so an integration logging in
	// as "base" would sign the user out of the web app; integrations use a platform of
	// their own, which a Sugar administrator registers under Admin > Configure API
	// Platforms. Every instance ships with a public OAuth key
	// named "sugar" that has no secret, which is used unless the connection supplies its
	// own client credentials. See "/oauth2/token POST" in the Sugar Developer Guide.
	SetInfo(SugarCRM, ProviderInfo{
		DisplayName: "SugarAI",
		AuthType:    Oauth2,
		BaseURL:     "{{.workspace}}/rest/v11_24",
		Oauth2Opts: &Oauth2Opts{
			GrantType: Password,
			TokenURL:  "{{.workspace}}/rest/v11_24/oauth2/token",
			TokenParams: map[string]string{
				"client_id":     "sugar",
				"client_secret": "",
				"platform":      "{{.platform}}",
			},
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348686/media/sugarCRM_1722348686.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348716/media/sugarCRM_1722348716.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348686/media/sugarCRM_1722348686.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348716/media/sugarCRM_1722348716.svg",
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
					Name:        "workspace",
					DisplayName: "Base URL",
					Prompt: "The address of your Sugar instance, without a trailing slash. " +
						"For example, https://mycompany.sugarondemand.com.",
				},
				{
					Name:        "platform",
					DisplayName: "API Platform",
					DocsURL:     "https://support.sugarai.com/documentation/sugar_versions/14.0/serve/administration_guide/developer_tools/", //nolint:lll
					Prompt: "The API platform name that your Sugar administrator added under " +
						"Admin > Configure API Platforms for this integration. " +
						"Do not use base, which would sign the user out of Sugar.",
				},
			},
		},
	})
}
