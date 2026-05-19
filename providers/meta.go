package providers

import "github.com/amp-labs/connectors/common"

const (
	Meta           Provider        = "meta"
	ModuleFacebook common.ModuleID = "facebook"
)

func init() {
	SetInfo(Meta, ProviderInfo{
		DisplayName: "Meta",
		AuthType:    Oauth2,
		BaseURL:     "https://graph.facebook.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.facebook.com/v23.0/dialog/oauth",
			TokenURL:                  "https://graph.facebook.com/v23.0/oauth/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
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
		DefaultModule: ModuleFacebook,
		Modules: &Modules{
			ModuleFacebook: {
				BaseURL:     "https://graph.facebook.com",
				DisplayName: "Facebook",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098801/media/meta.com_1753098801.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098836/media/meta.com_1753098836.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098801/media/meta.com_1753098801.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1753098858/media/meta.com_1753098858.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "adAccountId",
					DisplayName: "Ad account id",
				},
				{
					Name:        "businessId",
					DisplayName: "Business id",
				},
			},
		},
	})
}
