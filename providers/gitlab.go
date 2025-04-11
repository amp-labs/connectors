package providers

import "github.com/amp-labs/connectors/common"

const GitLab Provider = "gitLab"

func init() {
	SetInfo(GitLab, ProviderInfo{
		DisplayName: "GitLab",
		AuthType:    Oauth2,
		BaseURL:     "https://gitlab.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://gitlab.com/oauth/authorize",
			TokenURL:                  "https://gitlab.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Modules: &Modules{
			common.ModuleRoot: {
				BaseURL:     "https://gitlab.com/api/v4",
				DisplayName: "GitLab",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003317/media/GitLab_1734003316.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003260/media/GitLab_1734003258.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003317/media/GitLab_1734003316.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1734003350/media/GitLab_1734003349.svg",
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
}
