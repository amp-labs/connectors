package providers

import "github.com/amp-labs/connectors/common"

const LinkedIn Provider = "linkedIn"

const (
	// ModulePlatform is the module used for regular linkedIn objects.
	ModulePlatform common.ModuleID = "platform"
	// ModuleAds is the module used for ads related objects.
	ModuleAds common.ModuleID = "ads"
)

func init() {
	// LinkedIn configuration
	SetInfo(LinkedIn, ProviderInfo{
		DisplayName: "LinkedIn",
		AuthType:    Oauth2,
		BaseURL:     "https://api.linkedin.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.linkedin.com/oauth/v2/authorization",
			TokenURL:                  "https://www.linkedin.com/oauth/v2/accessToken",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DefaultModule: ModulePlatform,
		Modules: &Modules{
			ModulePlatform: {
				BaseURL:     "https://api.linkedin.com",
				DisplayName: "LinkedIn (Regular)",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleAds: {
				BaseURL:     "https://api.linkedin.com",
				DisplayName: "LinkedIn (Ads)",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225364/media/c2esjc2pb5o1qa9bwi0b.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724225364/media/c2esjc2pb5o1qa9bwi0b.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722481059/media/linkedIn_1722481058.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722481017/media/linkedIn_1722481016.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "adAccountId",
					DisplayName: "Ad Account ID",
					DocsURL:     "https://www.linkedin.com/help/linkedin/answer/a424270/find-linkedin-ads-account-details",
					ModuleDependencies: &ModuleDependencies{
						ModuleAds: ModuleDependency{},
					},
				},
			},
		},
	})
}
