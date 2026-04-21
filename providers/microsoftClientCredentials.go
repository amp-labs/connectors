package providers

import "net/http"

// MicrosoftClientCredentials is a twin of the Microsoft provider that
// authenticates using the OAuth2 client credentials grant instead of the
// per-user Authorization Code grant. It targets the same Microsoft Graph
// APIs, so the connector implementation in providers/microsoft is reused
// under a different provider name.
//
// Use cases include admin-consented bulk access (accessing many users'
// mailboxes/calendars without individual OAuth flows), background services,
// and any scenario where the app acts as itself rather than on behalf of a
// signed-in user.
const MicrosoftClientCredentials Provider = "microsoftClientCredentials"

func init() {
	SetInfo(MicrosoftClientCredentials, ProviderInfo{
		DisplayName: "Microsoft",
		AuthType:    Oauth2,
		BaseURL:     "https://graph.microsoft.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://graph.microsoft.com/v1.0/organization",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://login.microsoftonline.com/{{.workspace}}/oauth2/v2.0/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: true,
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
					Name:        "workspace",
					DisplayName: "Tenant ID",
					Prompt:      "The Azure AD tenant GUID (e.g. `951a1899-8810-4356-ax10-3a5f8fg99a65`)",
					DocsURL:     "https://docs.withampersand.com/customer-guides/microsoft-client-credentials",
				},
			},
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
