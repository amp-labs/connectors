package providers

const Okta Provider = "okta"

//nolint:lll
func init() {
	SetInfo(Okta, ProviderInfo{
		DisplayName: "Okta",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.okta.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.okta.com/oauth2/v1/authorize",
			TokenURL:                  "https://{{.workspace}}.okta.com/oauth2/v1/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			DocsURL:                   "https://developer.okta.com/docs/guides/implement-oauth-for-okta/main/#get-an-access-token-and-make-a-request",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1759883320/media/okta_1759883320.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1759883372/media/okta_1759883372.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1759883280/media/okta_1759883279.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1759883387/media/okta_1759883386.svg",
			},
		},
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true, // Proxy testing completed
			Read:      true, // Read support with Link header pagination
			Subscribe: false,
			Write:     true,  // Write support for user provisioning (create/update)
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Domain",
					DocsURL:     "https://developer.okta.com/docs/guides/find-your-domain/main/",
				},
			},
		},
	})
}
