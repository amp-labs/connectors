package providers

const Notion Provider = "notion"

func init() {
	// Notion Configuration
	SetInfo(Notion, ProviderInfo{
		DisplayName: "Notion",
		AuthType:    Oauth2,
		BaseURL:     "https://api.notion.com",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722167069/media/notion.com_1722167068.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722167069/media/notion.com_1722167068.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722167069/media/notion.com_1722167068.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722167069/media/notion.com_1722167068.jpg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.notion.com/v1/oauth/authorize",
			TokenURL:                  "https://api.notion.com/v1/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
			TokenMetadataFields: TokenMetadataFields{
				ConsumerRefField:  "owner.user.id",
				WorkspaceRefField: "workspace_id",
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
