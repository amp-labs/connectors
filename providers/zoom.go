package providers

import "github.com/amp-labs/connectors/common"

const Zoom Provider = "zoom"

const (
	ModuleZoomUser    common.ModuleID = "user"
	ModuleZoomMeeting common.ModuleID = "meeting"
)

func init() {
	// Zoom configuration
	SetInfo(Zoom, ProviderInfo{
		DisplayName: "Zoom",
		AuthType:    Oauth2,
		BaseURL:     "https://api.zoom.us",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://zoom.us/oauth/authorize",
			TokenURL:                  "https://zoom.us/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DefaultModule: ModuleZoomMeeting,
		Modules: &Modules{
			ModuleZoomUser: {
				BaseURL:     "https://api.zoom.us/v2",
				DisplayName: "Zoom (User)",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleZoomMeeting: {
				BaseURL:     "https://api.zoom.us/v2",
				DisplayName: "Zoom (Meeting)",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325775/media/const%20Zoom%20Provider%20%3D%20%22zoom%22_1722325765.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325874/media/const%20Zoom%20Provider%20%3D%20%22zoom%22_1722325874.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325775/media/const%20Zoom%20Provider%20%3D%20%22zoom%22_1722325765.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722325900/media/const%20Zoom%20Provider%20%3D%20%22zoom%22_1722325899.svg",
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
