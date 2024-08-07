package providers

const RingCentral Provider = "ringCentral"

func init() {
	// RingCentral configuration
	SetInfo(RingCentral, ProviderInfo{
		DisplayName: "RingCentral",
		AuthType:    Oauth2,
		BaseURL:     "https://platform.ringcentral.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 PKCE,
			AuthURL:                   "https://platform.ringcentral.com/restapi/oauth/authorize",
			TokenURL:                  "https://platform.ringcentral.com/restapi/oauth/token",
			ExplicitScopesRequired:    false,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "owner_id",
			},
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470246/media/ringCentral_1722470246.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470263/media/ringCentral_1722470262.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470246/media/ringCentral_1722470246.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722470278/media/ringCentral_1722470278.svg",
			},
		},
	})
}
