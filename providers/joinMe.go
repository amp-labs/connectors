package providers

const JoinMe Provider = "joinMe"

func init() {
	// JoinMe configuration
	SetInfo(JoinMe, ProviderInfo{
		DisplayName: "Join Me",
		AuthType:    Oauth2,
		BaseURL:     "https://api.join.me",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://secure.join.me/api/public/v1/auth/oauth2",
			TokenURL:                  "https://secure.join.me/api/public/v1/auth/token",
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1739958047/media/api.join.me_1739958046.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739958019/media/api.join.me_1739958017.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1739958047/media/api.join.me_1739958046.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1739958019/media/api.join.me_1739958017.png",
			},
		},
	})
}
