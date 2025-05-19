package providers

import "github.com/amp-labs/connectors/common"

const Marketo Provider = "marketo"

const (
	// ModuleMarketoAssets is the module/API used for accessing assets objects.
	ModuleMarketoAssets common.ModuleID = "assets"
	// ModuleMarketoLeads is the module/API used for accessing leads objects.
	ModuleMarketoLeads common.ModuleID = "leads"
)

// nolint:funlen
func init() {
	// Marketo configuration file
	// workspace maps to marketo instance
	SetInfo(Marketo, ProviderInfo{
		DisplayName: "Marketo",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.mktorest.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328319/media/marketo_1722328318.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328291/media/marketo_1722328291.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328319/media/marketo_1722328318.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328291/media/marketo_1722328291.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			TokenURL:                  "https://{{.workspace}}.mktorest.com/identity/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			GrantType:                 ClientCredentials,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DefaultModule: ModuleMarketoLeads,
		Modules: &Modules{
			ModuleMarketoAssets: {
				BaseURL:     "https://{{.workspace}}.mktorest.com/asset/v1",
				DisplayName: "Marketo (Assets)",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleMarketoLeads: {
				BaseURL:     "https://{{.workspace}}.mktorest.com/v1",
				DisplayName: "Marketo (Leads)",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
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
			Read:      true,
			Subscribe: false,
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name: "workspace",
				},
			},
		},
	})
}
