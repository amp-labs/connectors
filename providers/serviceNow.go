package providers

const (
	ServiceNow     Provider = "serviceNow"
	ServiceNowPKCE Provider = "serviceNowPKCE"
)

func init() { //nolint:lll
	// ServiceNow configuration
	SetInfo(ServiceNow, ProviderInfo{
		DisplayName: "ServiceNow",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.service-now.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://{{.workspace}}.service-now.com/oauth_auth.do",
			TokenURL:                  "https://{{.workspace}}.service-now.com/oauth_token.do",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/tn5f6xh2nbb3bops7bsn.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/tn5f6xh2nbb3bops7bsn.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405162/media/const%20ServiceNow%20Provider%20%3D%20%22serviceNow%22_1722405162.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405162/media/const%20ServiceNow%20Provider%20%3D%20%22serviceNow%22_1722405162.svg",
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
			Search: SearchSupport{
				Operators: SearchOperators{
					Equals: true,
				},
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Instance",
				},
			},
		},
	})

	// ServiceNow configuration
	SetInfo(ServiceNowPKCE, ProviderInfo{
		DisplayName: "ServiceNow",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.service-now.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://{{.workspace}}.service-now.com/oauth_auth.do",
			TokenURL:                  "https://{{.workspace}}.service-now.com/oauth_token.do",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 AuthorizationCodePKCE,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/tn5f6xh2nbb3bops7bsn.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/tn5f6xh2nbb3bops7bsn.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405162/media/const%20ServiceNow%20Provider%20%3D%20%22serviceNow%22_1722405162.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722405162/media/const%20ServiceNow%20Provider%20%3D%20%22serviceNow%22_1722405162.svg",
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
			Search: SearchSupport{
				Operators: SearchOperators{
					Equals: false,
				},
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Instance",
				},
			},
		},
	})
}
