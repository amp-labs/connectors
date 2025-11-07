package providers

import "net/http"

const Calendly Provider = "calendly"

func init() {
	// Calendly Configuration
	SetInfo(Calendly, ProviderInfo{
		DisplayName: "Calendly",
		AuthType:    Oauth2,
		BaseURL:     "https://api.calendly.com",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.calendly.com/users/me",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://auth.calendly.com/oauth/authorize",
			TokenURL:                  "https://auth.calendly.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		PostAuthInfoNeeded: true,
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346654/media/calendly_1722346653.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346580/media/calendly_1722346580.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724365119/media/gzqssdg62nudhokl9sms.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722346580/media/calendly_1722346580.svg",
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
			PostAuthentication: []MetadataItemPostAuthentication{
				{
					Name: "UserURI",
				},
				{
					Name: "OrganizationURI",
				},
			},
		},
	})
}
