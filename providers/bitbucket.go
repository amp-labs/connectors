package providers

const Bitbucket Provider = "bitbucket"

func init() {
	SetInfo(Bitbucket, ProviderInfo{
		DisplayName: "Bitbucket",
		AuthType:    Oauth2,
		BaseURL:     "https://api.bitbucket.org",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://bitbucket.org/site/oauth2/authorize",
			TokenURL:                  "https://bitbucket.org/site/oauth2/access_token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741019187/media/bitbucket.org_1741019186.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741019214/media/bitbucket.org_1741019213.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741019187/media/bitbucket.org_1741019186.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1741019214/media/bitbucket.org_1741019213.svg",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					// similar to organization in Github
					Name:        "workspace",
					DisplayName: "Workspace",
				},
			},
		},
	})
}
