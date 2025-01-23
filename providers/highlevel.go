package providers

const (
	HighLevelStandard   Provider = "highLevelStandard"
	HighLevelWhiteLabel Provider = "highLevelWhiteLabel"
)

//nolint:funlen
func init() {
	// HighlevelStandard configuration
	SetInfo(HighLevelStandard, ProviderInfo{
		DisplayName: "Highlevel Standard",
		AuthType:    Oauth2,
		BaseURL:     "https://services.leadconnectorhq.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://marketplace.gohighlevel.com/oauth/chooselocation",
			TokenURL:                  "https://services.leadconnectorhq.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "userId",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624760/media/gohighlevel.com_1737624759.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624575/media/gohighlevel.com_1737624573.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624760/media/gohighlevel.com_1737624759.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624627/media/gohighlevel.com_1737624627.png",
			},
		},
	})

	// HighlevelWhiteLabel configuration
	SetInfo(HighLevelWhiteLabel, ProviderInfo{
		DisplayName: "Highlevel White Label",
		AuthType:    Oauth2,
		BaseURL:     "https://services.leadconnectorhq.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://marketplace.leadconnectorhq.com/oauth/chooselocation",
			TokenURL:                  "https://services.leadconnectorhq.com/oauth/token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField:      "scope",
				ConsumerRefField: "userId",
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624760/media/gohighlevel.com_1737624759.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624575/media/gohighlevel.com_1737624573.png",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624760/media/gohighlevel.com_1737624759.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1737624627/media/gohighlevel.com_1737624627.png",
			},
		},
	})
}
