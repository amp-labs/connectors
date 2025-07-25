// nolint:lll
package providers

const (
	DynamicsBusinessCentral Provider = "dynamicsBusinessCentral"
	DynamicsCRM             Provider = "dynamicsCRM"
)

func init() { // nolint:funlen
	// Microsoft Dynamics 365 Business Central configuration
	SetInfo(DynamicsBusinessCentral, ProviderInfo{
		DisplayName: "Microsoft Dynamics Business Central",
		AuthType:    Oauth2,
		BaseURL:     "https://api.businesscentral.dynamics.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/authorize",
			TokenURL:                  "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/eajuugwekqardkcwf45c.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346267/media/dynamicsCRM_1722346267.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346298/media/dynamicsCRM_1722346297.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346267/media/dynamicsCRM_1722346267.svg",
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
					DisplayName: "Tenant ID",
					DocsURL:     "https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/tenant-admin-center",
				},
				{
					Name:        "companyId",
					DisplayName: "Company ID",
					DocsURL:     "https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/api-reference/v2.0/api/dynamics_company_get",
				},
				{
					Name:        "environmentName",
					DisplayName: "Environment Name",
					DocsURL:     "https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/administration/tenant-admin-center-environments",
				},
			},
		},
	})

	// Dynamics CRM Configuration
	SetInfo(DynamicsCRM, ProviderInfo{
		DisplayName: "Microsoft Dynamics CRM",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.api.crm.dynamics.com/api/data",
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:               "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			ExplicitScopesRequired: true,
			// TODO: flip this to false once we implement the ability to get workspace
			// information post-auth.
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/eajuugwekqardkcwf45c.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346267/media/dynamicsCRM_1722346267.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346298/media/dynamicsCRM_1722346297.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346267/media/dynamicsCRM_1722346267.svg",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Organization ID",
				},
			},
		},
	})
}
