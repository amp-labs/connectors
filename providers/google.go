package providers

import "github.com/amp-labs/connectors/common"

const (
	Google Provider = "google"
)

const (
	// ModuleGoogleAds is the module used for listing Ad campaigns and groups.
	ModuleGoogleAds common.ModuleID = "ads"
	// ModuleGoogleCalendar is the module used for listing user calendars.
	ModuleGoogleCalendar common.ModuleID = "calendar"
	// ModuleGoogleContacts is the module used for listing contacts from People API.
	ModuleGoogleContacts common.ModuleID = "contacts"
)

//nolint:funlen
func init() {
	SetInfo(Google, ProviderInfo{
		DisplayName: "Google",
		AuthType:    Oauth2,
		BaseURL:     "https://www.googleapis.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.google.com/o/oauth2/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline", "prompt": "consent"},
			TokenURL:                  "https://oauth2.googleapis.com/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		DefaultModule: ModuleGoogleCalendar,
		Modules: &Modules{
			ModuleGoogleAds: {
				BaseURL:     "https://googleads.googleapis.com",
				DisplayName: "Google Ads",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
			ModuleGoogleCalendar: {
				BaseURL:     "https://www.googleapis.com/calendar",
				DisplayName: "Google Calendar",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleGoogleContacts: {
				BaseURL:     "https://people.googleapis.com",
				DisplayName: "Google Contacts",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349084/media/google_1722349084.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349053/media/google_1722349052.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349084/media/google_1722349084.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722349053/media/google_1722349052.svg",
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					// Google Ads authorization requires Developer Token.
					Name:        "developerToken",
					DisplayName: "Developer Token",
					DocsURL:     "https://developers.google.com/google-ads/api/rest/auth#developer_token",
					ModuleDependencies: &ModuleDependencies{
						ModuleGoogleAds: ModuleDependency{},
					},
				},
				{
					// Google Ads API calls are done by a manager of customers.
					Name:        "loginCustomerId",
					DisplayName: "Login Customer Id (Google Ads Manager Id)",
					DocsURL:     "https://developers.google.com/google-ads/api/rest/auth#login_customer_id",
					ModuleDependencies: &ModuleDependencies{
						ModuleGoogleAds: ModuleDependency{},
					},
				},
				{
					// Google Ads API call must have a customer as a context to query their Campaigns, Ads.
					Name:        "customerId",
					DisplayName: "Customer Id",
					DocsURL:     "https://developers.google.com/google-ads/api/rest/auth#login_customer_id",
					ModuleDependencies: &ModuleDependencies{
						ModuleGoogleAds: ModuleDependency{},
					},
				},
			},
		},
	})
}
