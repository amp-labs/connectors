package providers

const Drift Provider = "drift"

func init() {
	// Drift Configuration
	SetInfo(Drift, ProviderInfo{
		DisplayName: "Drift",
		AuthType:    Oauth2,
		BaseURL:     "https://driftapi.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://dev.drift.com/authorize",
			TokenURL:                  "https://driftapi.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "orgId",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722448523/media/drift_1722448523.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722448401/media/drift_1722448400.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722448486/media/drift_1722448485.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722448371/media/drift_1722448370.svg",
			},
		},
	})
}
