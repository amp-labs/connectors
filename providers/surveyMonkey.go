package providers

const SurveyMonkey Provider = "surveyMonkey"

func init() {
	// SurveyMonkey configuration file
	SetInfo(SurveyMonkey, ProviderInfo{
		DisplayName: "SurveyMonkey",
		AuthType:    Oauth2,
		BaseURL:     "https://api.surveymonkey.com",
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://api.surveymonkey.com/oauth/authorize",
			TokenURL:                  "https://api.surveymonkey.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064886/media/surveyMonkey_1722064885.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064863/media/surveyMonkey_1722064862.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064941/media/surveyMonkey_1722064939.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722064920/media/surveyMonkey_1722064919.svg",
			},
		},
	})
}
