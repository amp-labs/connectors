package providers

const (
	Dropbox     Provider = "dropbox"
	DropboxSign Provider = "dropboxSign"
)

//nolint:all
func init() {
	// Dropbox configuration
	SetInfo(Dropbox, ProviderInfo{
		DisplayName: "Dropbox",
		AuthType:    Oauth2,
		BaseURL:     "https://api.dropboxapi.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://www.dropbox.com/oauth2/authorize",
			TokenURL:                  "https://api.dropboxapi.com/oauth2/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "account_id",
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724223403/media/qoxime3z8bloqgzsvude.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722492220/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722492221.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722491962/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722491963.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722492197/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722492198.svg",
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
	})

	SetInfo(DropboxSign, ProviderInfo{
		DisplayName: "Dropbox Sign",
		AuthType:    Oauth2,
		BaseURL:     "https://api.hellosign.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://app.hellosign.com/oauth/authorize",
			TokenURL:                  "https://app.hellosign.com/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722491962/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722491963.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722492220/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722492221.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722491962/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722491963.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722492197/media/Dropbox%20%20%20%20%20Provider%20%3D%20%22dropbox%22_1722492198.svg",
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
	})
}
