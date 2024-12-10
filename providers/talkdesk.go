package providers

const Talkdesk Provider = "talkdesk"

func init() {
	SetInfo(Talkdesk, ProviderInfo{
		DisplayName: "Talkdesk",
		AuthType:    Oauth2,
		BaseURL:     "https://api.talkdeskapp.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://{{.workspace}}.talkdeskid.com/oauth/authorize",
			TokenURL:                  "https://{{.workspace}}.talkdeskid.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: true,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733431983/media/talkdesk.com_1733431982.png",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1733432333/media/talkdesk.com_1733432332.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733431983/media/talkdesk.com_1733431982.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733432426/media/talkdesk.com_1733432426.svg",
			},
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
	})
}
