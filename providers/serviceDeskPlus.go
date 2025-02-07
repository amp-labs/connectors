package providers

const ServiceDeskPlusUS Provider = "serviceDeskPlusUS"

// ServiceDesk Plus is hosted at multiple data centers, and therefore available on different domains.
//
//nolint:lll
func init() {
	// ServiceDesk Plus US configuration
	SetInfo(ServiceDeskPlusUS, ProviderInfo{
		DisplayName: "ServiceDesk Plus US",
		AuthType:    Oauth2,
		BaseURL:     "https://sdpondemand.manageengine.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://accounts.zoho.com/oauth/v2/auth",
			AuthURLParams:             map[string]string{"access_type": "offline"},
			TokenURL:                  "https://accounts.zoho.com/oauth/v2/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			DocsURL:                   "https://www.manageengine.com/products/service-desk/sdpod-v3-api/getting-started/oauth-2.0.html#register-your-application",
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738753124/media/manageengine.com_1738753122.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738655830/media/servicedeskplus.com_1738655829.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738753124/media/manageengine.com_1738753122.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738655830/media/servicedeskplus.com_1738655829.svg",
			},
		},
	})
}
