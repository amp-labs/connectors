package providers

const (
	Google         Provider = "google"
	GoogleContacts Provider = "googleContacts"
	GoogleAds      Provider = "googleAds"
)

//nolint:funlen
func init() {
	// Google Support Configuration
	SetInfo(Google, ProviderInfo{
		DisplayName: "Google",
		AuthType:    Oauth2,
		BaseURL:     "https://www.googleapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://oauth2.googleapis.com/token",
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
	})

	// GoogleContacts Support Configuration
	SetInfo(GoogleContacts, ProviderInfo{
		DisplayName: "Google Contacts",
		AuthType:    Oauth2,
		BaseURL:     "https://people.googleapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://oauth2.googleapis.com/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})

	// GoogleAds Support Configuration
	SetInfo(GoogleAds, ProviderInfo{
		DisplayName: "Google Ads",
		AuthType:    Oauth2,
		BaseURL:     "https://googleads.googleapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://oauth2.googleapis.com/token",
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
			Proxy:     true,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
	})
}
