package providers

const Coupa Provider = "coupa"

func init() {
	// Coupa configuration
	SetInfo(Coupa, ProviderInfo{
		DisplayName: "Coupa",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}/oauth2/authorizations/new?",
			TokenURL:                  "https://{{.workspace}}/oauth/token",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770299248/media/coupa.com_1770299248.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770299226/media/coupa.com_1770299225.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770299299/media/coupa.com_1770299298.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1770299268/media/coupa.com_1770299267.svg",
			},
		},

		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Instance domain",
					DocsURL:     "https://compass.coupa.com/en-us/products/product-documentation/integration-technical-documentation/the-coupa-core-api/get-started-with-the-api", //nolint:lll
				},
			},
		},
	})
}
