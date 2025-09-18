package providers

import "github.com/amp-labs/connectors/common"

const (
	Seismic Provider = "seismic"

	// ModuleReporting is the module used for accessing and manging reporting API.
	ModuleReporting common.ModuleID = "reporting"

	// ModuleIntegration is the module used for accessing and manging integration API.
	ModuleIntegration common.ModuleID = "integration"
)

func init() { // nolint: funlen
	SetInfo(Seismic, ProviderInfo{
		DisplayName: "Seismic",
		AuthType:    Oauth2,
		BaseURL:     "https://api.seismic.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.seismic.com/tenants/{{.workspace}}/connect/authorize",
			TokenURL:                  "https://auth.seismic.com/tenants/{{.workspace}}/connect/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348404/media/seismic_1722348404.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348429/media/seismic_1722348428.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348404/media/seismic_1722348404.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722348448/media/seismic_1722348447.svg",
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
		DefaultModule: ModuleReporting,
		Modules: &Modules{
			ModuleReporting: {
				DisplayName: "Seismic Reporting",
				BaseURL:     "https://api.seismic.com/reporting",
				Support: Support{
					Proxy: false,
					Read:  false,
					Write: false,
				},
			},
			ModuleIntegration: {
				DisplayName: "Seismic Integration",
				BaseURL:     "https://api.seismic.com/integration",
				Support: Support{
					Proxy: false,
					Read:  false,
					Write: false,
				},
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Tenant",
				},
			},
		},
	})
}
