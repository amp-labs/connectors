package providers

const LinkedIn Provider = "linkedIn"

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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722481032/media/linkedIn_1722481031.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722481000/media/linkedIn_1722480999.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722481059/media/linkedIn_1722481058.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722481017/media/linkedIn_1722481016.svg",
			},
		},
		Labels: &Labels{
			LabelKeyClusterID: LabelValueClusterID,
		},
	})
}
