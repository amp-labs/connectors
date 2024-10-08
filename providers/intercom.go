package providers

const Intercom Provider = "intercom"

func init() {
	// Intercom configuration
	SetInfo(Intercom, ProviderInfo{
		DisplayName: "Intercom",
		AuthType:    Oauth2,
		BaseURL:     "https://api.intercom.io",
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/zscxf6amk8pu2ijejrw0.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327671/media/intercom_1722327670.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724364085/media/srib8u1d8vgtik0j2fww.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722327671/media/intercom_1722327670.svg",
			},
		},
		Oauth2Opts: &Oauth2Opts{
			AuthURL:                   "https://app.intercom.com/oauth",
			TokenURL:                  "https://api.intercom.io/auth/eagle/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			GrantType:                 AuthorizationCode,
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
