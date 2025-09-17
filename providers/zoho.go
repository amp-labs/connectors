package providers

const Zoho Provider = "zoho"

func init() {
	// Zoho configuration
	SetInfo(Zoho, ProviderInfo{
		DisplayName: "Zoho",
		AuthType:    Oauth2,
		// E.g. www.zohoapis.com, www.zohoapis.eu, www.zohoapis.in, etc.
		BaseURL:            "https://{{.zoho_api_domain}}",
		PostAuthInfoNeeded: true,
		Metadata: &ProviderMetadata{
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "zoho_api_domain",
				},
				{
					Name: "zoho_token_domain",
				},
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType: AuthorizationCode,
			// NB: This works for all Zoho regions (com, eu, in, cn, au, etc). It will redirect
			// to the appropriate domain based on the user's account. It's ok to hard-code
			// the .com domain here. And since we don't know the user's region in advance,
			// we can't use a templated domain like in BaseURL and TokenURL.
			AuthURL: "https://accounts.zoho.com/oauth/v2/auth",
			// ref: https://www.zoho.com/analytics/api/v2/authentication/generating-code.html
			AuthURLParams: map[string]string{"access_type": "offline"},
			// E.g. accounts.zoho.com, accounts.zoho.eu, accounts.zoho.in, etc.
			TokenURL:                  "https://{{.zoho_token_domain}}/oauth/v2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				WorkspaceRefField: "api_domain",
				ScopesField:       "scope",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724224295/media/lk7ohfgtmzys1sl919c8.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471872/media/zoho_1722471871.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471890/media/zoho_1722471890.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722471890/media/zoho_1722471890.svg",
			},
		},
	})
}
