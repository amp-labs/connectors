package providers

import "net/http"

// MicrosoftWorkspaceDelegation is a twin of the Microsoft provider that
// authenticates using OAuth2 client credentials with admin-consented
// application permissions, instead of the per-user Authorization Code grant.
// It targets the same Microsoft Graph APIs, so the connector implementation
// in providers/microsoft is reused under a different provider name.
//
// Motivation: the Authorization Code flow requires each end user to complete
// an OAuth consent flow individually. Admin consent lets a tenant admin
// approve the app once, and the platform can then access any user's data
// (mail, calendar, contacts) in that tenant via client credentials — no
// per-user OAuth flows needed.
const MicrosoftWorkspaceDelegation Provider = "microsoftWorkspaceDelegation"

func init() {
	SetInfo(MicrosoftWorkspaceDelegation, ProviderInfo{
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
					Prompt:      "The Azure AD tenant GUID. The customer's admin can find this in Azure portal → Entra ID → Overview → Tenant ID. Admin consent must be granted for this tenant before creating a connection.",
					DocsURL:     "https://docs.withampersand.com/customer-guides/microsoft-workspace-delegation",
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
