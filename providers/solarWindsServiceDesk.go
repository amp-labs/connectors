package providers

const (
	SolarWindsServiceDesk Provider = "solarWindsServiceDesk"
)

// SolarWinds Service Desk has data centers in three regions: US, EU, and APJ, each with a different base URL.
//
//nolint:lll
func init() {
	// SolarWindsServiceDesk configuration
	SetInfo(SolarWindsServiceDesk, ProviderInfo{
		DisplayName: "SolarWinds Service Desk",
		AuthType:    ApiKey,
		BaseURL:     "https://{{.subdomain}}.samanage.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "X-Samanage-Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://documentation.solarwinds.com/en/success_center/swsd/content/completeguidetoswsd/token-authentication-for-api-integration.htm#link4",
		}, Support: Support{
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
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322100/media/solarwinds.com_1738322099.svg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1738322158/media/solarwinds.com_1738322159.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:         "subdomain",
					DisplayName:  "API Subdomain",
					DefaultValue: "api",
					DocsURL:      "https://apidoc.samanage.com/#section/General-Concepts/Service-URL",
					Prompt:       "Enter your region subdomain: api (US), apieu (Europe), or apiau (APJ)",
				},
			},
		},
	})
}
