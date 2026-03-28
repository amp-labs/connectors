package providers

import "github.com/amp-labs/connectors/common"

const (
	Google Provider = "google"
)

const (
	// ModuleGoogleCalendar is the module used for listing user calendars.
	// https://developers.google.com/workspace/calendar/api/v3/reference
	ModuleGoogleCalendar common.ModuleID = "calendar"
	// ModuleGoogleContacts is the module used for listing contacts from People API.
	// https://developers.google.com/people
	ModuleGoogleContacts common.ModuleID = "contacts"
	// ModuleGoogleGmail is the module used for listing emails from Gmail API.
	// https://developers.google.com/workspace/gmail/api/reference/rest
	ModuleGoogleGmail common.ModuleID = "gmail"
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
					Read:      true,
					Subscribe: false,
					Write:     true,
				},
			},
			ModuleGoogleGmail: {
				BaseURL:     "https://gmail.googleapis.com/gmail",
				DisplayName: "Gmail",
				Support: Support{
					Read:      true,
					Subscribe: false,
					Write:     true,
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
	})
}
