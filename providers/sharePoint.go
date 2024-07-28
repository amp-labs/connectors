package providers

const SharePoint Provider = "sharePoint"

func init() {
	SetInfo(SharePoint, ProviderInfo{
		DisplayName: "SharePoint",
		AuthType:    Oauth2,
		BaseURL:     "https://{{.workspace}}.sharepoint.com/_api",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166950/media/microsoft.com_1722166949.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166950/media/microsoft.com_1722166949.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166950/media/microsoft.com_1722166949.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722166950/media/microsoft.com_1722166949.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:              AuthorizationCode,
			AuthURL:                "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:               "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			ExplicitScopesRequired: true,
			// TODO: Switch to post-auth metadata collection
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
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
