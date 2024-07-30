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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346320/media/dynamicsCRM_1722346320.jpg",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346320/media/dynamicsCRM_1722346320.jpg",
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
	})
}
