package providers

const Outreach Provider = "outreach"

func init() {
	// Outreach Configuration
	SetInfo(Outreach, ProviderInfo{
		DisplayName: "Outreach",
		AuthType:    Oauth2,
		BaseURL:     "https://api.outreach.io",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329361/media/outreach_1722329360.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329335/media/outreach_1722329335.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329361/media/outreach_1722329360.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722329335/media/outreach_1722329335.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.outreach.io/oauth/authorize",
			TokenURL:                  "https://api.outreach.io/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
	})
}
