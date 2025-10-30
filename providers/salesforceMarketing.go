package providers

const SalesforceMarketing Provider = "salesforceMarketing"

func init() {
	// SalesforceMarketing configuration
	SetInfo(SalesforceMarketing, ProviderInfo{
		DisplayName: "Salesforce Marketing Cloud",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.rest.marketingcloudapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://{{.workspace}}.auth.marketingcloudapis.com/v2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Subdomain",
					DocsURL:     "https://help.salesforce.com/s/articleView?language=en_US&id=sf.faq_domain_name_what.htm&type=5",
				},
				{
					Name: "inputShouldNotBeCollectedForSalesforceMarketing",
					ModuleDependencies: &ModuleDependencies{
						ModuleOtherModule: ModuleDependency{},
					},
				},
				{
					Name: "inputShouldBeCollectedForSalesforceMarketing",
				},
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470590/media/salesforce_1722470589.svg",
			},
		},
	})
}
