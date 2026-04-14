package providers

// GoogleWorkspaceDelegation is a twin of the Google provider that authenticates using
// a Google Workspace service account with domain-wide delegation instead of
// the Authorization Code grant. It targets the same underlying Google
// Workspace APIs and modules (Calendar, Contacts, Gmail), so the connector
// implementation in providers/google is reused under a different provider name.
//
// Motivation: the Authorization Code flow requires each end user to complete
// an OAuth consent flow individually. Domain-wide delegation lets a Workspace
// admin authorize a service account once, and the platform can then access
// any user's data in the domain by impersonating them via JWT assertions —
// no per-user OAuth flows needed.
const GoogleWorkspaceDelegation Provider = "googleWorkspaceDelegation"

//nolint:funlen
func init() {
	SetInfo(GoogleWorkspaceDelegation, ProviderInfo{
		DisplayName: "Google (Domain-Wide Delegation)",
		AuthType:    Custom,
		BaseURL:     "https://www.googleapis.com",
		CustomOpts: &CustomAuthOpts{
			// Token acquisition is handled by the server's DynamicHeadersGenerator:
			// service account JSON → JWT with sub=userEmail → token endpoint → Bearer header.
			Inputs: []CustomAuthInput{
				{
					Name:        "serviceAccountKey",
					DisplayName: "Service Account Key (Base64)",
					Prompt:      "Base64-encoded JSON key file for a GCP service account with domain-wide delegation enabled.",
					DocsURL:     "https://docs.withampersand.com/provider-guides/google-workspace-delegation",
				},
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
					Read:      false,
					Subscribe: false,
					Write:     false,
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "userEmail",
					DisplayName: "User Email",
					Prompt:      "The Google Workspace email address of the user to access (e.g. `user@company.com`). The service account will impersonate this user via domain-wide delegation.",
					DocsURL:     "https://docs.withampersand.com/provider-guides/google-workspace-delegation",
				},
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
