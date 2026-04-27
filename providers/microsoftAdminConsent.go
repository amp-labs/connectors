package providers

import "net/http"

// MicrosoftAdminConsent is a twin of the Microsoft provider that authenticates
// using OAuth2 client credentials after an Azure AD admin grants tenant-wide
// consent for application permissions. It targets the same Microsoft Graph
// APIs, so the connector implementation in providers/microsoft is reused
// under a different provider name.
//
// The setup phase uses the /adminconsent endpoint (redirect-based, like auth
// code) to obtain tenant admin approval. The runtime phase uses the
// client_credentials grant with the ProviderApp's credentials to get
// app-only tokens that can access any user's data in the consented tenant.
//
// See https://learn.microsoft.com/en-us/entra/identity-platform/v2-admin-consent
// See https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-client-creds-grant-flow
const MicrosoftAdminConsent Provider = "microsoftAdminConsent"

func init() {
	SetInfo(MicrosoftAdminConsent, ProviderInfo{
		DisplayName: "Microsoft (Admin consent)",
		AuthType:    Oauth2,
		BaseURL:     "https://graph.microsoft.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://graph.microsoft.com/v1.0/organization",
		},
		Oauth2Opts: &Oauth2Opts{
			// GrantType is AuthorizationCode so the UI library routes to the
			// popup OAuth flow (OauthFlow2) and /v1/oauth-connect generates
			// the auth URL. The AuthURL points to Microsoft's admin consent
			// endpoint instead of the standard /authorize endpoint. At runtime,
			// the connection uses client_credentials for token acquisition
			// (AUTH_SCHEME_OAUTH2_CLIENT is set explicitly in the callback).
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://login.microsoftonline.com/organizations/v2.0/adminconsent",
			TokenURL:                  "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false, // workspace (tenant ID) comes from the admin consent redirect
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328808/media/microsoft_1722328808.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328785/media/microsoft_1722328785.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328808/media/microsoft_1722328808.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722328785/media/microsoft_1722328785.svg",
			},
		},
	})
}
