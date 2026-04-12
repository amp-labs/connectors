package providers

import (
	"github.com/amp-labs/connectors/common"
)

const (
	Clio               Provider        = "clio"
	ModuleClioPlatform common.ModuleID = "platform"
	ModuleClioManage   common.ModuleID = "manage"
)

func init() {
	SetInfo(Clio, ProviderInfo{
		DisplayName: "Clio",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.clio_api_domain}}",
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			// Integrations with Clio products are local to the regional instance.
			// It's not possible, for example, create a US integration and have it work for Canadian Clio customers.
			//
			// Manage OAuth(US):
			//   https://app.clio.com/oauth/authorize
			//   https://app.clio.com/oauth/token
			// Platform OAuth(US):
			//   https://auth.api.clio.com/oauth/authorize
			//   https://auth.api.clio.com/oauth/token
			AuthURL:                   "https://{{.clio_oauth_domain}}/oauth/authorize",
			TokenURL:                  "https://{{.clio_oauth_domain}}/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://docs.developers.clio.com/handbook/getting-started/get-a-developer-account",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776024411/media/clio.com_1776024411.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776024397/media/clio.com_1776024397.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776024373/media/clio.com_1776024372.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776024349/media/clio.com_1776024349.svg",
			},
		},
		// The Clio Platform is the current platform for building on Clio Grow.
		// In future, it will be the single point of entry for building integrations with all Clio products.
		// Doc: https://docs.developers.clio.com/handbook/getting-started/clio-manage-and-clio-platform/#multi-product-integrations
		DefaultModule: ModuleClioManage,
		Modules: &Modules{
			ModuleClioPlatform: {
				BaseURL:     "https://{{.clio_api_domain}}",
				DisplayName: "Clio Platform",
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
				BaseURL:     "https://{{.clio_api_domain}}",
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
				// Required for both Manage and Platform.
				// Examples: Manage US app.clio.com, Platform US api.clio.com (see Clio regional docs).
				{
					Name:         "clio_api_domain",
					DisplayName:  "Clio API domain",
					DefaultValue: "app.clio.com",
					DocsURL:      "https://docs.developers.clio.com/handbook/getting-started/regions",
					Prompt:       "Provide Clio Regional API domain(Manage/Platform). e.g. Manage US: app.clio.com, Platform US: api.clio.com",
					ModuleDependencies: &ModuleDependencies{
						ModuleClioPlatform: ModuleDependency{},
						ModuleClioManage:   ModuleDependency{},
					},
				},
				// Platform uses a distinct auth host (e.g. US auth.api.clio.com). Manage uses the same host as the API;
				{
					Name:         "clio_oauth_domain",
					DisplayName:  "Clio OAuth domain",
					DefaultValue: "app.clio.com",
					DocsURL:      "https://docs.developers.clio.com/handbook/getting-started/regions",
					Prompt:       "Provide Clio Regional OAuth domain(Manage/Platform). e.g. Manage US: app.clio.com, Platform US: auth.api.clio.com",
					ModuleDependencies: &ModuleDependencies{
						ModuleClioPlatform: ModuleDependency{},
						ModuleClioManage:   ModuleDependency{},
					},
				},
			},
		},
	})
}
