package providers

import "net/http"

const HelpScoutMailbox Provider = "helpScoutMailbox"

func init() {
	// HelpScoutMailbox Support Configuration
	SetInfo(HelpScoutMailbox, ProviderInfo{
		DisplayName: "Help Scout Mailbox",
		AuthType:    Oauth2,
		BaseURL:     "https://api.helpscout.net",
		AuthHealthCheck: &AuthHealthCheck{
			Method:             http.MethodGet,
			SuccessStatusCodes: []int{http.StatusOK},
			Url:                "https://api.helpscout.net/v2/users/me",
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 ClientCredentials,
			TokenURL:                  "https://api.helpscout.net/v2/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://developer.helpscout.com/mailbox-api/overview/authentication/",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722061926/media/helpScoutMailbox_1722061925.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722061868/media/helpScoutMailbox_1722061867.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722061926/media/helpScoutMailbox_1722061925.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722061899/media/helpScoutMailbox_1722061898.svg",
			},
		},
	})
}
