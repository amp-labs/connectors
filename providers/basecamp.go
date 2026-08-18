package providers

const Basecamp Provider = "basecamp"

func init() {
	// Basecamp Configuration.
	// The workspace in baseURL maps to accounts[].id from the authorization document.
	// Every Basecamp resource endpoint is prefixed with the account ID; only
	// https://launchpad.37signals.com/authorization.json can be reached without it,
	// which is where a user can look the account ID up if the URL is not to hand.
	SetInfo(Basecamp, ProviderInfo{
		DisplayName: "Basecamp",
		AuthType:    Oauth2,
		BaseURL:     "https://3.basecampapi.com/{{.workspace}}",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324615/media/basecamp_1722324614.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324674/media/basecamp_1722324673.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324615/media/basecamp_1722324614.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722324674/media/basecamp_1722324673.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://launchpad.37signals.com/authorization/new",
			TokenURL:                  "https://launchpad.37signals.com/authorization/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
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
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Account ID",
					Prompt:      "our account ID is the unique number located in your Basecamp URL. For example, in `https://app.basecamp.com/6258233/projects`, the account ID is 6258233.`", // nolint:lll
				},
			},
		},
	})
}
