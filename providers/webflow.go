package providers

const Webflow Provider = "webflow"

func init() {
	// Webflow Support Configuration
	SetInfo(Webflow, ProviderInfo{
		DisplayName: "Webflow",
		AuthType:    Oauth2,
		BaseURL:     "https://api.webflow.com",
		Oauth2Opts: &Oauth2Opts{
			GrantType:                 AuthorizationCode,
			AuthURL:                   "https://webflow.com/oauth/authorize",
			TokenURL:                  "https://api.webflow.com/oauth/access_token",
			ExplicitScopesRequired:    true,
			ExplicitWorkspaceRequired: false,
			TokenMetadataFields: TokenMetadataFields{
				ScopesField: "scope",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					// Every object except `sites` is scoped to a site
					// (/v2/sites/{site_id}/...). The connector substitutes this
					// value into the request path.
					Name:        "siteId",
					DisplayName: "Site ID",
					DocsURL:     "https://help.webflow.com/hc/en-us/sections/33776399764499-Site-Settings",
					Prompt:      "The ID of the Webflow site to connect (e.g. `6ac4a4402b7737869ac6f61d`), found under Site settings > General.", //nolint:lll
				},
			},
		},
		//nolint:lll
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1724169124/media/uzen6vmatu35qsrc3zry.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347649/media/const%20Webflow%20Provider%20%3D%20%22webflow%22_1722347650.jpg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347649/media/const%20Webflow%20Provider%20%3D%20%22webflow%22_1722347650.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1722347433/media/const%20Webflow%20Provider%20%3D%20%22webflow%22_1722347433.svg",
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
