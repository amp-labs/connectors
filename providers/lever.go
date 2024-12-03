package providers

const (
	Lever        Provider = "lever"
	LeverSandbox Provider = "leverSandbox"
)

func init() {
	// Lever Production configuration
	SetInfo(Lever, ProviderInfo{

		DisplayName: "Lever",
		AuthType:    Oauth2,
		BaseURL:     "https://api.lever.co",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.lever.co/authorize",
			TokenURL:                  "https://auth.lever.co/oauth/token",
			Audience:                  []string{"https://api.lever.co/v1/"},
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733231877/media/lever.co_1733231842.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733232008/media/lever.co_1733231986.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733231877/media/lever.co_1733231842.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733231953/media/lever.co_1733231938.svg",
			},
		},
	})

	// Lever Sandbox configuration
	SetInfo(LeverSandbox, ProviderInfo{

		DisplayName: "Lever Sandbox",
		AuthType:    Oauth2,
		BaseURL:     "https://api.sandbox.lever.co",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://sandbox-lever.auth0.com/authorize",
			TokenURL:                  "https://sandbox-lever.auth0.com/oauth/token",
			Audience:                  []string{"https://api.sandbox.lever.co/v1/"},
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733231877/media/lever.co_1733231842.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733232008/media/lever.co_1733231986.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733231877/media/lever.co_1733231842.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1733231953/media/lever.co_1733231938.svg",
			},
		},
	})
}
