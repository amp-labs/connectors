package providers

const AdobeExperiencePlatform Provider = "adobeExperiencePlatform"

func init() {
	// AdobeExperiencePlatform 2-legged auth
	SetInfo(AdobeExperiencePlatform, ProviderInfo{
		AuthType: Oauth2,
		BaseURL:  "https://platform.adobe.io",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://ims-na1.adobelogin.com/ims/token/v3",
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
	})
}
