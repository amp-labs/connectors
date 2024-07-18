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
	})
}
