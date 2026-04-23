package providers

const (
	QuickBooks        Provider = "quickbooks"
	QuickbooksSandbox Provider = "quickbooksSandbox"
)

func init() { //nolint:funlen
	SetInfo(QuickBooks, ProviderInfo{
		DisplayName: "QuickBooks",
		AuthType:    Oauth2,
		BaseURL:     "https://quickbooks.api.intuit.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://appcenter.intuit.com/connect/oauth2",
			TokenURL:               "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer",
			ExplicitScopesRequired: true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753941999/media/quickbooks.com_1753941998.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753942143/media/quickbooks.com_1753942142.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753941999/media/quickbooks.com_1753941998.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753942143/media/quickbooks.com_1753942142.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "realmId",
					DisplayName: "Company ID",
					DocsURL:     "https://coda.io/@leandro-zubrezki/quickbooks-pack-start-here/find-your-qbo-company-id-15", // nolint:lll
				},
			},
		},
	})

	SetInfo(QuickbooksSandbox, ProviderInfo{
		DisplayName: "QuickBooks SandBox",
		AuthType:    Oauth2,
		BaseURL:     "https://sandbox-quickbooks.api.intuit.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://appcenter.intuit.com/connect/oauth2",
			TokenURL:               "https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer",
			ExplicitScopesRequired: true,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753941999/media/quickbooks.com_1753941998.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753942143/media/quickbooks.com_1753942142.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753941999/media/quickbooks.com_1753941998.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753942143/media/quickbooks.com_1753942142.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "realmId",
					DisplayName: "Company ID",
					DocsURL:     "https://coda.io/@leandro-zubrezki/quickbooks-pack-start-here/find-your-qbo-company-id-15", // nolint:lll
				},
			},
		},
	})
}
