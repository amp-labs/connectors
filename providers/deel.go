package providers

const (
	Deel        Provider = "deel"
	DeelSandbox Provider = "deelSandbox"
)

// nolint: funlen
func init() {
	// Deel configuration
	SetInfo(Deel, ProviderInfo{
		DisplayName: "Deel",
		AuthType:    Oauth2,
		BaseURL:     "https://api.letsdeel.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.deel.com/oauth2/authorize",
			TokenURL:                  "https://app.deel.com/oauth2/tokens",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://developer.deel.com/api/oauth",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774977960/media/deel.com_1774977958.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774977989/media/deel.com_1774977988.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774977960/media/deel.com_1774977958.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774978008/media/deel.com_1774978007.svg",
			},
		},
	})

	// Deel Sandbox configuration
	SetInfo(DeelSandbox, ProviderInfo{
		DisplayName: "Deel Sandbox",
		AuthType:    Oauth2,
		BaseURL:     "https://api-sandbox.demo.deel.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.demo.deel.com/oauth2/authorize",
			TokenURL:                  "https://app.demo.deel.com/oauth2/tokens",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://developer.deel.com/api/oauth",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774977960/media/deel.com_1774977958.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774977989/media/deel.com_1774977988.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774977960/media/deel.com_1774977958.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1774978008/media/deel.com_1774978007.svg",
			},
		},
	})
}
