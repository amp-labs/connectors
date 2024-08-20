package providers

const Discord Provider = "discord"

func init() {
	// Discord Support Configuration
	SetInfo(Discord, ProviderInfo{
		DisplayName: "Discord",
		AuthType:    Oauth2,
		BaseURL:     "https://discord.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://discord.com/oauth2/authorize",
			TokenURL:                  "https://discord.com/api/oauth2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		//nolint:all
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/y9kecrrv3wtlzbfihfjh.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169123/media/y9kecrrv3wtlzbfihfjh.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722323153/media/const%20Discord%20Provider%20%3D%20%22discord%22_1722323153.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722323175/media/const%20Discord%20Provider%20%3D%20%22discord%22_1722323174.svg",
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
