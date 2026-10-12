package providers

import (
	"github.com/amp-labs/connectors/common"
)

const (
	Clio             Provider        = "clio"
	ModuleClioGrow   common.ModuleID = "grow"
	ModuleClioManage common.ModuleID = "manage"
)

const clioRegionTemplate = "{{if and .region (ne .region \"us\") (ne .region \"US\")}}{{.region}}.{{end}}"

func init() { //nolint:funlen
	SetInfo(Clio, ProviderInfo{
		DisplayName: "Clio",
		AuthType:    Oauth2,
		BaseURL:     "https://" + clioRegionTemplate + "{{.workspace}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			// Integrations with Clio products are local to the regional instance.
			// It's not possible, for example, to create a US integration and have it work for Canadian Clio
			// customers.
			//
			// Manage OAuth(US):
			//   https://app.clio.com/oauth/authorize
			//   https://app.clio.com/oauth/token
			// Platform OAuth(US):
			//   https://auth.api.clio.com/oauth/authorize
			//   https://auth.api.clio.com/oauth/token
			//
			// Region is a hostname prefix (eu, ca, au). Some environments provide "us" for the US region;
			// treat it as empty (no prefix).
			//
			// Manage OAuth uses the app host; Platform OAuth uses auth.api.clio.com (not api.clio.com).
			AuthURL: "https://" + clioRegionTemplate +
				"{{if eq .workspace \"api.clio.com\"}}auth.api.clio.com{{else}}{{.workspace}}{{end}}" +
				"/oauth/authorize",
			TokenURL: "https://" + clioRegionTemplate +
				"{{if eq .workspace \"api.clio.com\"}}auth.api.clio.com{{else}}{{.workspace}}{{end}}" +
				"/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			DocsURL: "https://docs.developers.clio.com/handbook/getting-started/" +
				"get-a-developer-account",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/" +
					"v1776024411/media/clio.com_1776024411.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/" +
					"v1776024397/media/clio.com_1776024397.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/" +
					"v1776024373/media/clio.com_1776024372.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/" +
					"v1776024349/media/clio.com_1776024349.svg",
			},
		},
		// The Clio Platform is the current platform for building on Clio Grow.
		// In future, it will be the single point of entry for building integrations with all Clio products.
		// Doc: https://docs.developers.clio.com/handbook/getting-started/clio-manage-and-clio-platform
		DefaultModule: ModuleClioManage,
		Modules: &Modules{
			ModuleClioGrow: {
				BaseURL:     "https://" + clioRegionTemplate + "api.clio.com",
				DisplayName: "Clio Grow",
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
			},
			ModuleClioManage: {
				BaseURL:     "https://" + clioRegionTemplate + "app.clio.com",
				DisplayName: "Clio Manage",
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
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:         "region",
					DisplayName:  "Region",
					DefaultValue: "",
					DocsURL: "https://docs.developers.clio.com/handbook/getting-started/" +
						"regions",
					Prompt: "Regional hostname prefix: eu, ca, or au. Leave empty for US.",
					ModuleDependencies: &ModuleDependencies{
						ModuleClioGrow:   ModuleDependency{},
						ModuleClioManage: ModuleDependency{},
					},
				},
				{
					Name:         "workspace",
					DisplayName:  "API host",
					DefaultValue: "app.clio.com",
					DocsURL: "https://docs.developers.clio.com/handbook/getting-started/" +
						"regions",
					Prompt: "Clio Manage: \"app.clio.com\". Clio Grow: \"api.clio.com\".",
					ModuleDependencies: &ModuleDependencies{
						ModuleClioGrow:   ModuleDependency{},
						ModuleClioManage: ModuleDependency{},
					},
				},
			},
		},
	})
}
