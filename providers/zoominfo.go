package providers

const ZoomInfo Provider = "zoominfo"

func init() {
	// ZoomInfo configuration
	SetInfo(ZoomInfo, ProviderInfo{
		DisplayName: "ZoomInfo",
		AuthType:    Oauth2,
		BaseURL:     "https://api.zoominfo.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.zoominfo.com/gtm/oauth/v1/authorize",
			TokenURL:                  "https://api.zoominfo.com/gtm/oauth/v1/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCodePKCE,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621074/media/zoominfo.com_1758621078.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621183/media/zoominfo.com_1758621188.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621074/media/zoominfo.com_1758621078.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1758621167/media/zoominfo.com_1758621172.svg",
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
